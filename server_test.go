package main

import (
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func TestAcceptConnections(t *testing.T) {
	ln, err := StartTCPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("StartTCPServer() error = %v", err)
	}
	defer ln.Close()

	go func() {
		_ = AcceptLoop(ln)
	}()

	for i := 0; i < 2; i++ {
		conn, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Fatalf("client %d failed to connect: %v", i+1, err)
		}
		conn.Close()
	}

	// Give the server a moment to accept both connections.
	time.Sleep(50 * time.Millisecond)
}

func TestHandlePeer(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	want := "SET name redis"

	done := make(chan error, 1)

	go func() {
		done <- HandlePeer(server)
	}()

	_, err := client.Write([]byte(want))
	if err != nil {
		t.Fatalf("client write failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandlePeer() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HandlePeer() timed out")
	}

	_ = io.EOF
	_ = os.Stdout
}
