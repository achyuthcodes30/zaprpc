package zaprpc

import (
	"context"
	"errors"
	"github.com/quic-go/quic-go"
	"io"
	"net"
)

func isGracefulClose(err error) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, io.EOF),
		errors.Is(err, net.ErrClosed),
		errors.Is(err, context.Canceled):
		return true
	}
	var appErr *quic.ApplicationError
	if errors.As(err, &appErr) {
		return true
	}
	return false
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	type timeoutIface interface{ Timeout() bool }
	var te timeoutIface
	if errors.As(err, &te) && te.Timeout() {
		return true
	}
	var idle *quic.IdleTimeoutError
	return errors.As(err, &idle)
}
