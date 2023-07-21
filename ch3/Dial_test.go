package main

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// Create a listener on r random port
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 新建一个无缓冲通道
	done := make(chan struct{})
	// 新建一个协程
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			// 服务器端针对每一个TCP连接进行处理，在客户端关闭连接后
			// 会发送一个FIN包给服务器，服务器接收并读取到FIN包后，
			// 服务器对这个连接的读取方法`c.Read(buf)`会返回一个io.EOF错误
			// 这表示客户端已经关闭了连接。
			//服务器端的处理函数（即在 go func(c net.Conn) 中的函数）会在读取到
			//io.EOF 错误后退出循环，并在退出前调用 c.Close() 方法关闭连接。
			//这个操作会向客户端发送一个FIN包，表明服务器端也不再发送数据。
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("receive: %q", buf[:n])
				}
			}(conn)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	listener.Close()
	// 同步主Goroutine和处理连接的Goroutine，确保测试在所有处理都完成后才结束。
	<-done
}
