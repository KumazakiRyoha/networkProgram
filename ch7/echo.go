package echo

import (
	"context"
	"net"
)

func streamingEchoServer(ctx context.Context, network string,
	addr string) (net.Addr, error) {
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
		}()

		for {
			conn, err := s.Accept()
			if err != nil {
				return
			}

			go func() {
				defer func() { _ = conn.Close() }()

				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}

					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}

			}()
		}
	}()
	return s.Addr(), nil
}