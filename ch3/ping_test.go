package ch3

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingerAdvanceDeadline(t *testing.T) {

	// 创建了一个 done 通道，类型为 struct{}。struct{} 是 Go 中的一个空结构体，
	// 通常用于信号传递。在这个测试函数中，done 通道被用来表示 Pinger goroutine 何时完成。
	done := make(chan struct{})
	// 创建了一个 TCP 监听器，它在本地监听一个系统分配的端口。
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// 在 Go 语言中，time.Now() 函数返回当前时间，类型为 time.Time。
	begin := time.Now()
	go func() {
		defer func() { close(done) }()
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			err := conn.Close()
			if err != nil {
				return
			}
		}()
		resetTimer := make(chan time.Duration, 1)
		// 这两行代码创建了一个 resetTimer 通道，并向通道发送了一个 1
		// 秒的时间间隔。resetTimer 通道被传递给 Pinger 函数，用于指定 "ping" 消息的发送间隔。
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		//这行代码为连接 conn 设置了一个读写操作的截止时间。
		//如果到了这个时间点，连接上的读写操作还没有完成，那么这些操作就会返回一个错误
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		// 这两行代码创建了一个缓冲区，然后在一个无限循环中从连接上
		// 读取数据到这个缓冲区。每次读取到数据后，我们都会打印已经过去的时间和接收到的数据
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

			// 我们向 resetTimer 通道发送了一个 0 秒的时间间隔，这会让 Pinger 使用之前的时间间隔继续发送
			// "ping" 消息。同时，我们为连接设置了一个新的截止时间，这个新的截止时间是当前时间加上 5 秒。
			// 这样，我们就有另外 5 秒的时间来接收下一个 "ping" 消息。
			resetTimer <- 0
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	// 这行代码创建了一个到本地监听器的TCP连接。net.Dial
	// 函数返回一个代表这个连接的 net.Conn 对象和一个错误对象
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// goroutine执行结束时关闭连接
	defer conn.Close()
	// 这两行代码创建了一个缓冲区，然后在一个循环中从连接上读取数据到这个缓冲区。
	buf := make([]byte, 1024)
	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	_, err = conn.Write([]byte("PONG!!!")) // should reset the ping timer
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			if err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				break
			}
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	// 通道被关闭了。在这种情况下，接收操作会立即返回，并返回通道元素类型的零值。
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second {
		t.Fatalf("excepted EOF at 9 seconds;actual %s", end)
	}
}
