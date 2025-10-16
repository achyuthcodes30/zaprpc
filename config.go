package zaprpc

import (
	"crypto/tls"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)


type ServerConfig struct {
	TLSConfig       *tls.Config
	QUICConfig      *quic.Config
	QUICTransport *quic.Transport
	Logger *zap.Logger
}

type ClientConfig struct {
	TLSConfig       *tls.Config
	QUICConfig      *quic.Config
	Logger *zap.Logger
}
