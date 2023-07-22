package ch3

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {

	// 定义一个上下文，约定上下文在10秒后到期。同时创建一个可以取消该上下文的函数cancel
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// 启动一个goroutine 接受一个来自监听器的连接，然后立即关闭它。。
	go func() {
		// Only accepting a single connection
		conn, err := listener.Accept()
		if err == nil {
			err := conn.Close()
			if err != nil {
				return
			}
		}
	}()

	// 定义一个dial函数，这个函数会使用给定的上下文去拨打给定的地址，然后将其id发送到响应通道
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		err = c.Close()
		if err != nil {
			return
		}
		select {
		case <-ctx.Done():
		case response <- id:
		}
	}

	// 创建一个响应通道，一个等待组，用于等待所有的拨号尝试都完成
	res := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	response := <-res
	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context;actual: %s", ctx.Err())
	}
	t.Logf("dialer %d retrieved the resource", response)

}
