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
	Codec Codec
	Logger *zap.Logger

}

type ClientConfig struct {
	Codec Codec
	Logger *zap.Logger
}

type ConnectionConfig struct {
	TLSConfig       *tls.Config
	QUICConfig      *quic.Config
}
