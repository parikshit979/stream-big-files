package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/stream-big-files/transport"
)

type FileClient struct{}

func (fc *FileClient) Start(ctx context.Context, tType transport.TransportType, address, fileName, outputDir string) error {
	t := transport.NewTransport(tType, address)
	if t == nil {
		return fmt.Errorf("unsupported transport type: %v", tType)
	}

	handler := func(conn net.Conn) {
		defer conn.Close()
		if tType == transport.UDPTransportType {
			fc.handleUDPTransportConnReadLoop(conn, outputDir)
		} else {
			fc.handleTCPTransportConnReadLoop(conn, fileName, outputDir)
		}
	}

	// For client, Dial then call handler directly
	conn, err := t.Dial(ctx)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	handler(conn)
	return nil
}

func (fc *FileClient) handleTCPTransportConnReadLoop(conn net.Conn, fileName, outputDir string) {
	// Send request
	request := transport.TransportStreamRequest{FileName: fileName}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("failed to marshal stream request: %v", err)
	}
	if _, err := conn.Write(requestBytes); err != nil {
		log.Fatalf("failed to write to server: %v", err)
	}

	// Read file size
	var size int64
	if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
		log.Fatalf("failed to read file size: %v", err)
	}

	outPath := filepath.Join(outputDir, fileName)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	fmt.Printf("Receiving %s (%d bytes)...\n", fileName, size)
	written, err := io.CopyN(outFile, conn, size)
	if err != nil {
		log.Fatalf("failed to receive file: %v", err)
	}
	fmt.Printf("Received %d bytes, saved to %s\n", written, outPath)
}

func (fc *FileClient) handleUDPTransportConnReadLoop(conn net.Conn, outputDir string) {
	buf := make([]byte, 65535)
	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		log.Fatal("failed to type cast to UDPConn")
	}
	outPath := filepath.Join(outputDir, "udp_output.bin")
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("UDP read error: %v", err)
		}
		if n > 0 {
			if _, err := outFile.Write(buf[:n]); err != nil {
				log.Fatalf("failed to write UDP data: %v", err)
			}
			fmt.Printf("Received %d bytes from %v (written to %s)\n", n, addr, outPath)
		}
	}
}
