package zaprpc

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/quic-go/quic-go"
	"net"
	"time"
	"errors"
)

func NewConn(ctx context.Context, target string, cfg *ClientConfig) (quic.Connection, error) {
	if target == "" {
		return nil, fmt.Errorf("empty target")
	}
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target %q: %w", target, err)
	}
	var clientCfg ClientConfig
	if cfg != nil {
		clientCfg = *cfg
	}
	if clientCfg.TLSConfig == nil {
		clientCfg.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS13,
			ServerName:         host,
			InsecureSkipVerify: false,
			NextProtos:         []string{"zaprpc"},
		}
	}

	if clientCfg.QUICConfig == nil {
		clientCfg.QUICConfig = &quic.Config{
			KeepAlivePeriod: 15 * time.Second,
		}
	}

	conn, err := quic.DialAddr(ctx, target, clientCfg.TLSConfig, clientCfg.QUICConfig)
	if err != nil {

		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	return conn, nil
}

func Zap(conn quic.Connection, serviceMethod string, args ...any) (any, error) {
	stream, err := conn.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()
	dec := gob.NewDecoder(stream)
	enc := gob.NewEncoder(stream)
	req := struct {
		ServiceMethod string
		Args          []any
	}{
		ServiceMethod: serviceMethod,
		Args:          args,
	}

	err = enc.Encode(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	var response ZapResponse
	err = dec.Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if err, ok := response.Value.(struct{ Error string }); ok && err.Error != "" {
		return nil, errors.New(err.Error)
	}

	return response.Value, nil
}

