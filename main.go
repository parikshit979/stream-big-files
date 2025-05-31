package main

import (
	"context"
	"time"

	"github.com/stream-big-files/client"
	"github.com/stream-big-files/server"
	"github.com/stream-big-files/transport"
)

func main() {
	go func() {
		time.Sleep(time.Second * 3)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c := &client.FileClient{}
		c.Start(ctx, transport.TCPTransportType, ":3000", "sample.txt", "./clientstore")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &server.FileServer{
		FileStoreDir: "./serverstore",
	}
	s.Start(ctx, transport.TCPTransportType, ":3000")
}
