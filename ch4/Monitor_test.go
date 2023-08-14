package main

import (
	"io"
	"log"
	"net"
	"os"
	"testing"
)

// Monitor embeds a log.Logger meant for logging network traffic
type Monitor struct {
	*log.Logger
}

// Write implements the io.Writer interface
func (m *Monitor) Write(p []byte) (int, error) {
	err := m.Output(2, string(p))
	if err != nil {
		log.Println(err)
	}
	return len(p), nil
}

func TestMonitor(t *testing.T) {
	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		b := make([]byte, 1024)
		// io.TeeReader 的作用是在从原始
		// io.Reader（这里是连接）读取数据的同时，
		// 将读取的数据写入另一个 io.Writer（这里是 Monitor）。
		r := io.TeeReader(conn, monitor)
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
		// io.MultiWriter方法能够将所有的写操作赋值到
		// 传入的所有Writer中
		w := io.MultiWriter(conn, monitor)
		_, err = w.Write(b[:n]) // echo the message
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}

	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	_ = conn.Close()
	<-done
}
