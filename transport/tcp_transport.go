package transport

import (
	"context"
	"log"
	"net"
	"time"
)

type TCPTransport struct {
	address string
}

func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{address: address}
}

func (t *TCPTransport) Dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "tcp", t.address)
}

func (t *TCPTransport) ListenAndServe(ctx context.Context, handler ConnHandler) error {
	ln, err := net.Listen("tcp", t.address)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("TCP listening on %s", t.address)

	for {
		ln.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
		conn, err := ln.Accept()
		select {
		case <-ctx.Done():
			log.Println("TCP server shutting down")
			return nil
		default:
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			log.Printf("TCP accept error: %v", err)
			continue
		}
		go func(c net.Conn) {
			defer t.Close(c)
			handler(c)
		}(conn)
	}
}

func (t *TCPTransport) Close(conn net.Conn) error {
	return conn.Close()
}
