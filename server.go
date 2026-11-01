package main

import (
	"fmt"
	"io"
	"net"
)

func StartTCPServer(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("The server is listening on Port :%v\n", ln.Addr())
	return ln, nil
}

func AcceptLoop(ln net.Listener) error {
	// Loop because i dont want to stop after connecting only one peer
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		if err := HandlePeer(conn); err != nil {
			conn.Close()
			return err
		}
		fmt.Printf("The Peer is Connected: %v\n", conn.RemoteAddr())
		conn.Close()
	}
}

func HandlePeer(conn net.Conn) error {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("%v", string(buf[:n]))
	}
}
