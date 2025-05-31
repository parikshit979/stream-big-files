package transport

import (
	"context"
	"io"
	"log"
	"net"
	"time"
)

type UDPTransport struct {
	address string
}

func NewUDPTransport(address string) *UDPTransport {
	return &UDPTransport{address: address}
}

func (t *UDPTransport) Dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "udp", t.address)
}

func (t *UDPTransport) ListenAndServe(ctx context.Context, handler ConnHandler) error {
	addr, err := net.ResolveUDPAddr("udp", t.address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("UDP listening on %s", t.address)

	buf := make([]byte, 65535)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		select {
		case <-ctx.Done():
			log.Println("UDP server shutting down")
			return nil
		default:
		}
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			log.Printf("UDP read error: %v", err)
			continue
		}
		// Handle each packet in a goroutine
		go func(data []byte, addr *net.UDPAddr) {
			packetConn := &udpPacketConn{
				UDPConn:    conn,
				remoteAddr: addr,
				data:       data,
			}
			handler(packetConn)
			packetConn.Close()
		}(append([]byte{}, buf[:n]...), remoteAddr)
	}
}

func (t *UDPTransport) Close(conn net.Conn) error {
	return conn.Close()
}

// udpPacketConn implements net.Conn for a single UDP packet
type udpPacketConn struct {
	*net.UDPConn
	remoteAddr *net.UDPAddr
	data       []byte
	readDone   bool
}

func (u *udpPacketConn) Read(b []byte) (int, error) {
	if u.readDone {
		return 0, io.EOF
	}
	copy(b, u.data)
	u.readDone = true
	return len(u.data), nil
}

func (u *udpPacketConn) Write(b []byte) (int, error) {
	return u.UDPConn.WriteToUDP(b, u.remoteAddr)
}
