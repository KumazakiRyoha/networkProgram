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

		for { // 持续地等待客户端连接，并为每一个连接启动一个新的协程进行处理
			conn, err := s.Accept()
			if err != nil {
				return
			}

			go func() {
				defer func() { _ = conn.Close() }()

				for { //这个循环持续地从当前客户端读取数据，并将接收到的数据回显到客户端
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
