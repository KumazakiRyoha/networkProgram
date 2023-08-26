package ch5

import (
	"context"
	"fmt"
	"net"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
		}()

		buf := make([]byte, 1024)

		for {
			// ReadFrom从数据包中将数据读取到buf
			n, clientAddr, err := s.ReadFrom(buf) // client to server
			if err != nil {
				return
			}
			// WriteTo将接收到的数据回写到源地址
			_, err = s.WriteTo(buf[:n], clientAddr) // server to client
			if err != nil {
				return
			}
		}
	}()
	return s.LocalAddr(), nil
}
