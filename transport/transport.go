package transport

import (
	"context"
	"net"
)

type TransportType string

const (
	TCPTransportType TransportType = "tcp"
	UDPTransportType TransportType = "udp"
	// Add WebSocketTransportType, HTTPTransportType as needed
)

type TransportStreamRequest struct {
	FileName string
}

type ConnHandler func(net.Conn)

type Transport interface {
	Dial(ctx context.Context) (net.Conn, error)
	ListenAndServe(ctx context.Context, handler ConnHandler) error
	Close(conn net.Conn) error
}

func NewTransport(transport TransportType, address string) Transport {
	switch transport {
	case TCPTransportType:
		return NewTCPTransport(address)
	case UDPTransportType:
		return NewUDPTransport(address)
	}
	return nil
}
