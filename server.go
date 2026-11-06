package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Ayushmangit/redisGO/protocol"
)

func StartTCPServer(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("The server is listening on Port :%v\n", ln.Addr())
	return ln, nil
}

func AcceptLoop(ln net.Listener, executor *Executor) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	clients := make(map[net.Conn]struct{})
	for {
		conn, err := ln.Accept()
		if err != nil {
			mu.Lock()
			for conn := range clients {
				_ = conn.Close()
			}
			mu.Unlock()
			wg.Wait()
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		mu.Lock()
		clients[conn] = struct{}{}
		mu.Unlock()

		fmt.Printf("The Peer is Connected: %v\n", conn.RemoteAddr())
		wg.Add(1)
		go func(conn net.Conn) {
			defer wg.Done()
			defer func() {
				mu.Lock()
				delete(clients, conn)
				mu.Unlock()
			}()
			defer conn.Close()

			if err := HandlePeer(conn, executor); err != nil {
				fmt.Printf("Error handling peer %v: %v\n", conn.RemoteAddr(), err)
			}
		}(conn)
	}
}

func HandlePeer(conn net.Conn, executor *Executor) error {
	decoder := protocol.NewDecoder(conn)
	for {
		value, err := decoder.Decode()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		command, err := ValueToCommand(value)
		if err != nil {
			res := NewError(err.Error())
			resBytes, err := res.Encode()
			if err != nil {
				return err
			}
			if _, err := conn.Write(resBytes); err != nil {
				return err
			}
			continue
		}
		response, err := executor.Execute(command)
		if err != nil {
			return err
		}
		resBytes, err := response.Encode()
		if err != nil {
			return err
		}
		if _, err := conn.Write(resBytes); err != nil {
			return err
		}
		fmt.Printf("%+v\n", response)

	}
}
