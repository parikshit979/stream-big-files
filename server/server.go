package server

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stream-big-files/transport"
)

type FileServer struct {
	FileStoreDir string
	wg           sync.WaitGroup
	cancel       context.CancelFunc
}

func (fs *FileServer) Start(ctx context.Context, tType transport.TransportType, address string) error {
	t := transport.NewTransport(tType, address)
	ctx, cancel := context.WithCancel(ctx)
	fs.cancel = cancel

	handler := func(conn net.Conn) {
		defer fs.wg.Done()
		if tType == transport.UDPTransportType {
			fs.handleUDPTransportConnWrite(conn)
		} else {
			fs.handleTCPTransportConnWrite(conn)
		}
	}

	fs.wg = sync.WaitGroup{}
	go func() {
		err := t.ListenAndServe(ctx, func(conn net.Conn) {
			fs.wg.Add(1)
			handler(conn)
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	fs.wg.Wait()
	return nil
}

func (fs *FileServer) Stop() {
	if fs.cancel != nil {
		fs.cancel()
	}
}

func sanitizeFileName(name string) (string, error) {
	name = filepath.Base(name)
	if strings.Contains(name, "..") {
		return "", errors.New("invalid file name")
	}
	return name, nil
}

func (fs *FileServer) handleTCPTransportConnWrite(conn net.Conn) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("TCP read error: %v", err)
		return
	}

	var req transport.TransportStreamRequest
	if err := json.Unmarshal(buf[:n], &req); err != nil {
		log.Printf("Invalid request: %v", err)
		return
	}

	fileName, err := sanitizeFileName(req.FileName)
	if err != nil {
		log.Printf("Sanitize error: %v", err)
		return
	}

	filePath := filepath.Join(fs.FileStoreDir, fileName)
	fi, err := os.Stat(filePath)
	if err != nil {
		log.Printf("File not found: %v", err)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Open error: %v", err)
		return
	}
	defer file.Close()

	if err := binary.Write(conn, binary.LittleEndian, fi.Size()); err != nil {
		log.Printf("Write size error: %v", err)
		return
	}
	written, err := io.CopyBuffer(conn, file, make([]byte, 64*1024))
	if err != nil {
		log.Printf("Copy error: %v", err)
		return
	}
	log.Printf("Sent %d bytes for %s", written, fileName)
}

func (fs *FileServer) handleUDPTransportConnWrite(conn net.Conn) {
	size := 1400
	data := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		log.Printf("UDP random read error: %v", err)
		return
	}
	n, err := conn.Write(data)
	if err != nil {
		log.Printf("UDP write error: %v", err)
		return
	}
	log.Printf("Sent %d bytes over UDP", n)
}
