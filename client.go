package zaprpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
	"net"
	"time"
)

type Client struct {
	codec  Codec
	logger *zap.Logger
}

func NewClient(cfg *ClientConfig) *Client {

	var clientCfg ClientConfig
	if cfg != nil {
		clientCfg = *cfg
	}
	if clientCfg.Logger == nil {
		clientCfg.Logger = zap.NewNop()
	}

	if clientCfg.Codec == nil {
		clientCfg.Codec = &GOBCodec{}
	}
	c := &Client{
		logger: clientCfg.Logger,
		codec:  clientCfg.Codec,
	}
	c.logger.Info("Client object created")
	return c
}

func NewConn(ctx context.Context, target string, cfg *ConnectionConfig) (*quic.Conn, error) {
	if target == "" {
		return nil, fmt.Errorf("empty target")
	}
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target %q: %w", target, err)
	}
	var connectionCfg ConnectionConfig
	if cfg != nil {
		connectionCfg = *cfg
	}
	if connectionCfg.TLSConfig == nil {
		connectionCfg.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS13,
			ServerName:         host,
			InsecureSkipVerify: false,
			NextProtos:         []string{"zaprpc"},
		}
	}

	if connectionCfg.QUICConfig == nil {
		connectionCfg.QUICConfig = &quic.Config{
			KeepAlivePeriod: 15 * time.Second,
		}
	}

	conn, err := quic.DialAddr(ctx, target, connectionCfg.TLSConfig, connectionCfg.QUICConfig)
	if err != nil {

		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	return conn, nil
}
func (c *Client) WithLogger(logger *zap.Logger) *Client {
	if logger != nil {
		c.logger = logger
	}
	return c
}

func (c *Client) WithCodec(codec Codec) *Client {
	if codec != nil {
		c.codec = codec
	}
	return c
}

func (c *Client) Codec() string {
	return c.codec.Name()
}

func (c *Client) Zap(conn *quic.Conn, serviceMethod string, args ...any) (any, error) {
	codec := c.codec
	logger := c.logger
	stream, err := conn.OpenStream()
	if err != nil {
		logger.Debug("Failed to open stream", zap.String("details", err.Error()))
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()
	req := struct {
		ServiceMethod string
		Args          []any
	}{
		ServiceMethod: serviceMethod,
		Args:          args,
	}

	err = codec.Marshal(stream, req)
	if err != nil {
		logger.Debug("Error encoding request", zap.String("details", err.Error()))
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	var resp ZapResponse
	err = codec.Unmarshal(stream, &resp)
	if err != nil {
		logger.Debug("Error decoding response", zap.String("details", err.Error()))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if err, ok := resp.Value.(struct{ Error string }); ok && err.Error != "" {
		return nil, errors.New(err.Error)
	}

	return resp.Value, nil
}
