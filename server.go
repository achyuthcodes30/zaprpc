package zaprpc
import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
	"math/big"
	"net"
	"reflect"
	"syscall"
	"time"
)

type Server struct {
	services map[string]any
	cfg      *ServerConfig
}

func NewServer(cfg *ServerConfig) *Server {
	var serverCfg ServerConfig
	if cfg != nil {
		serverCfg = *cfg
	}
	if serverCfg.Logger == nil {
		serverCfg.Logger = zap.NewNop()
	}
	if serverCfg.QUICConfig == nil {
		serverCfg.QUICConfig = &quic.Config{
			KeepAlivePeriod: 15 * time.Second,
		}
	}
	if serverCfg.TLSConfig == nil {
		serverCfg.TLSConfig = generateTLSConfig()
	}

	if serverCfg.Codec == nil {
		serverCfg.Codec = &GOBCodec{}
	}
	server := &Server{
		services: make(map[string]any),
		cfg:      &serverCfg,
	}
	server.cfg.Logger.Info("Server object created")
	return server
}

func NewTransport(addr string, logger *zap.Logger) (*quic.Transport, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	baseAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logger.Error("Invalid UDP address", zap.Error(err))
		return nil, fmt.Errorf("invalid UDP address %q: %w", addr, err)
	}

	port := baseAddr.Port
	if port == 0 {
		port = 6121
	}

	const maxRetries = 32

	for i := 0; i < maxRetries; i++ {
		tryPort := port + i
		tryAddr := &net.UDPAddr{
			IP:   baseAddr.IP,
			Port: tryPort,
		}

		conn, err := net.ListenUDP("udp", tryAddr)
		if err == nil {
			tr := &quic.Transport{Conn: conn}

			logger.Info("Created QUIC transport",
				zap.String("addr", tryAddr.String()),
			)

			return tr, nil
		}

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if errno, ok := opErr.Err.(syscall.Errno); ok && errno == syscall.EADDRINUSE {
				logger.Warn("Port busy, trying another port", zap.Int("port", tryPort))
				continue
			}
		}
		return nil, fmt.Errorf("failed to bind UDP %d: %w", tryPort, err)
	}
	ephemeral := &net.UDPAddr{IP: baseAddr.IP, Port: 0}
	udp, err := net.ListenUDP("udp", ephemeral)
	if err != nil {
		return nil, fmt.Errorf("failed ephemeral UDP bind: %w", err)
	}
	logger.Info("Created QUIC transport (ephemeral)", zap.String("addr", udp.LocalAddr().String()))
	return &quic.Transport{Conn: udp}, nil
}

func (s *Server) WithTransport(transport *quic.Transport) *Server {

	if transport != nil {
		s.cfg.QUICTransport = transport
	}
	return s
}
func (s *Server) WithQUICConfig(quicConfig *quic.Config) *Server {

	if quicConfig != nil {
		s.cfg.QUICConfig = quicConfig
	}
	return s
}
func (s *Server) WithTLSConfig(tlsConfig *tls.Config) *Server {
	if tlsConfig != nil {
		s.cfg.TLSConfig = tlsConfig
	}
	return s
}
func (s *Server) WithLogger(logger *zap.Logger) *Server {
	if logger != nil {
		s.cfg.Logger = logger
	}
	return s
}

func (s *Server) RegisterService(name string, service any) {
	s.services[name] = service
	s.cfg.Logger.Info("Service added", zap.String("service-name", name))
}

func (s *Server) Serve(ctx context.Context) error {
	var (
		tr           *quic.Transport
		ln           *quic.Listener
		err          error
		ownTransport bool
	)
	cfg := s.cfg
	logger := cfg.Logger
	if cfg.QUICTransport != nil {
		tr = cfg.QUICTransport
	} else {
		tr, err = NewTransport(":6121", logger)
		if err != nil {
			logger.Error("UDP transport creation error", zap.Error(err))
			return err
		}
		ownTransport = true
	}

	defer func() {
		if ownTransport && tr != nil {
			_ = tr.Close()
		}
	}()
	ln, err = tr.Listen(cfg.TLSConfig, cfg.QUICConfig)
	if err != nil {
		logger.Error("Listener failure", zap.Error(err))
		return err
	}
	defer ln.Close()

	logger.Info("Server listening", zap.String("addr", ln.Addr().String()))
	for {
		conn, err := ln.Accept(ctx)
		if err != nil {
			switch {
			case isGracefulClose(err) || ctx.Err() != nil:
				logger.Info("Server shutting down")
				return nil
			case isTimeout(err):
				logger.Debug("Server timeout", zap.String("details", err.Error()))
				return nil
			default:
				logger.Error("Accept failed", zap.Error(err))
				return err
			}
		}
		go s.handleSession(ctx, conn)
	}

}

func (s *Server) handleSession(ctx context.Context, conn quic.Connection) {
	logger := s.cfg.Logger
	defer func() { _ = conn.CloseWithError(0, "Server closing") }()
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			switch {

			case isGracefulClose(err) || ctx.Err() != nil:
				logger.Info("Session closed", zap.String("details", err.Error()))
				return
			case isTimeout(err):
				logger.Debug("Session timeout", zap.String("details", err.Error()))
				return
			default:
				logger.Error("Error accepting stream", zap.Error(err))
				return
			}
		}
		go s.handleStream(ctx, stream)
	}
}

func (s *Server) handleStream(ctx context.Context, stream quic.Stream) {
	logger := s.cfg.Logger
	codec := s.cfg.Codec
	defer stream.Close()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Stream panic", zap.Any("recover", r), zap.Stack("stack"))
			stream.CancelRead(0)
			stream.CancelWrite(0)
		}
	}()
	var req struct {
		ServiceMethod string
		Args          []any
	}
	err := codec.Unmarshal(stream,&req)
	if err != nil {
		switch {
		case isGracefulClose(err) || ctx.Err() != nil:
			logger.Debug("Stream closed by peer", zap.String("details", err.Error()))
		case isTimeout(err):
			logger.Debug("Stream timeout", zap.String("details", err.Error()))
		default:
			logger.Warn("Decode failed; closing stream", zap.String("details", err.Error()))
		}
		stream.CancelRead(0)
		stream.CancelWrite(0)
		return
	}

	resp, err := s.callMethod(req.ServiceMethod, req.Args)
	if err != nil {
		logger.Error("Error calling method", zap.Error(err))
		if e := codec.Marshal(stream,ZapResponse{Value:struct{ Error string }{err.Error()}}); e != nil {
			logger.Error("Error encoding error reply", zap.Error(e))
			stream.CancelWrite(0)
			return
		}
		return
	}
	err = codec.Marshal(stream,ZapResponse{Value: resp})
	if err != nil {
		logger.Error("Error encoding response", zap.Error(err))
		stream.CancelWrite(0)
		return
	}
}

func (s *Server) callMethod(serviceMethod string, args []any) (any, error) {
	serviceName, methodName, found := parseServiceMethod(serviceMethod)
	if !found {
		s.cfg.Logger.Error("Invalid service method")
		return nil, fmt.Errorf("invalid service method: %s", serviceMethod)
	}

	service, ok := s.services[serviceName]
	if !ok {
		s.cfg.Logger.Error("Service not found")
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	method := reflect.ValueOf(service).MethodByName(methodName)
	if !method.IsValid() {
		s.cfg.Logger.Error("Method not found")
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflectArgs[i] = reflect.ValueOf(arg)
	}

	results := method.Call(reflectArgs)

	// If the method returns an error, it should be the last return value
	if len(results) > 0 {
		lastResult := results[len(results)-1]
		if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !lastResult.IsNil() {
				return nil, lastResult.Interface().(error)
			}
			results = results[:len(results)-1] // Remove the error from results
		}
	}

	// If there's only one result (excluding a potential error), return it directly
	if len(results) == 1 {
		return results[0].Interface(), nil
	}

	// If there are multiple results, return them as a slice
	response := make([]interface{}, len(results))
	for i, result := range results {
		response[i] = result.Interface()
	}

	return response, nil
}

func parseServiceMethod(serviceMethod string) (string, string, bool) {
	for i := 0; i < len(serviceMethod); i++ {
		if serviceMethod[i] == '.' {
			return serviceMethod[:i], serviceMethod[i+1:], true
		}
	}
	return "", "", false
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"zaprpc"},
	}
}
