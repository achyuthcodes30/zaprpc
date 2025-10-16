package zaprpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"net"
	"time"
	"errors"
	"go.uber.org/zap"
)


type Client struct{
	Codec Codec
	Logger *zap.Logger
}

func NewClient(cfg *ClientConfig) *Client{

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
		Logger: clientCfg.Logger,
		Codec: clientCfg.Codec,
	}
	c.Logger.Info("Client object created")
	return c
}

func NewConn(ctx context.Context, target string, cfg *ConnectionConfig) (quic.Connection, error) {
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

func (c *Client) Zap(ctx context.Context, conn quic.Connection, serviceMethod string, args ...any) (any, error) {
	codec := c .Codec
	// logger := c.Logger
	stream, err := conn.OpenStream()
	if err != nil {
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

	err = codec.Marshal(stream,req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	var resp ZapResponse
	err = codec.Unmarshal(stream, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if err, ok := resp.Value.(struct{ Error string }); ok && err.Error != "" {
		return nil, errors.New(err.Error)
	}

	return resp.Value, nil
}

