package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Ayushmangit/redisGO/store/kv"
)

func TestAcceptConnections(t *testing.T) {
	store := kv.NewStore(1000, 0)
	ln, err := StartTCPServer("127.0.0.1:0")
	executor := NewExecutor(store, 1000)
	defer executor.Stop()
	if err != nil {
		t.Fatalf("StartTCPServer() error = %v", err)
	}
	defer ln.Close()

	go func() {
		_ = AcceptLoop(ln, executor)
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
	store := kv.NewStore(1000, 0)
	executor := NewExecutor(store, 1000)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan error, 1)

	go func() {
		done <- HandlePeer(server, executor)
	}()

	request := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nredis\r\n"

	writeDone := make(chan error, 1)
	go func() {
		_, err := client.Write([]byte(request))
		writeDone <- err
	}()

	reader := bufio.NewReader(client)

	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading response: %v", err)
	}

	if response != "+OK\r\n" {
		t.Fatalf("response = %q, want %q", response, "+OK\r\n")
	}

	if err := <-writeDone; err != nil {
		t.Fatalf("client write failed: %v", err)
	}

	value, err := store.Get("name")
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}

	if value != "redis" {
		t.Fatalf("store.Get() = %q, want %q", value, "redis")
	}

	if err := client.Close(); err != nil {
		t.Fatalf("client close failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandlePeer() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HandlePeer() timed out")
	}
}

func TestHandlePeerContinuesAfterInvalidCommand(t *testing.T) {
	store := kv.NewStore(1000, 0)
	executor := NewExecutor(store, 1000)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan error, 1)

	go func() {
		done <- HandlePeer(server, executor)
	}()

	requests := "*1\r\n$7\r\nUNKNOWN\r\n" +
		"*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nredis\r\n"

	writeDone := make(chan error, 1)

	go func() {
		_, err := client.Write([]byte(requests))
		writeDone <- err
	}()

	reader := bufio.NewReader(client)

	// Read the invalid command response.
	firstResponse, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading first response: %v", err)
	}

	if !strings.HasPrefix(firstResponse, "-unknown command") {
		t.Fatalf("unexpected first response: %q", firstResponse)
	}

	// Read the successful SET response.
	secondResponse, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading second response: %v", err)
	}

	if secondResponse != "+OK\r\n" {
		t.Fatalf("second response = %q, want %q", secondResponse, "+OK\r\n")
	}

	if err := <-writeDone; err != nil {
		t.Fatalf("writing requests: %v", err)
	}

	// Verify that SET actually modified the store.
	value, err := store.Get("name")
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}

	if value != "redis" {
		t.Fatalf("store.Get() = %q, want %q", value, "redis")
	}

	// Close the client so HandlePeer receives EOF.
	if err := client.Close(); err != nil {
		t.Fatalf("client close: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("HandlePeer() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HandlePeer() timed out")
	}
}

func TestAcceptLoopHandlesClientsConcurrently(t *testing.T) {
	store := kv.NewStore(1000, 0)
	executor := NewExecutor(store, 1000)

	ln, err := StartTCPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("StartTCPServer() error = %v", err)
	}
	defer ln.Close()

	go func() {
		_ = AcceptLoop(ln, executor)
	}()

	// Client 1 connects but remains idle.
	idleClient, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("idle client connection failed: %v", err)
	}
	defer idleClient.Close()

	// Client 2 connects and sends a SET command.
	activeClient, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("active client connection failed: %v", err)
	}
	defer activeClient.Close()

	_ = activeClient.SetDeadline(time.Now().Add(2 * time.Second))

	request := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nredis\r\n"

	if _, err := activeClient.Write([]byte(request)); err != nil {
		t.Fatalf("writing SET request: %v", err)
	}

	reader := bufio.NewReader(activeClient)

	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading SET response: %v", err)
	}

	if response != "+OK\r\n" {
		t.Fatalf("response = %q, want %q", response, "+OK\r\n")
	}

	value, err := store.Get("name")
	if err != nil {
		t.Fatalf("store.Get() error = %v", err)
	}

	if value != "redis" {
		t.Fatalf("store.Get() = %q, want %q", value, "redis")
	}
}

func TestAcceptLoopGracefulShutdown(t *testing.T) {
	store := kv.NewStore(1000, 0)
	executor := NewExecutor(store, 1000)
	defer executor.Stop()

	ln, err := StartTCPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("StartTCPServer() error = %v", err)
	}

	done := make(chan error, 1)

	go func() {
		done <- AcceptLoop(ln, executor)
	}()

	// Allow AcceptLoop to start before closing the listener.
	time.Sleep(20 * time.Millisecond)

	if err := ln.Close(); err != nil {
		t.Fatalf("listener close failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("AcceptLoop() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("AcceptLoop() did not shut down in time")
	}
}
