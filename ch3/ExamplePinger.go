package ch3

import (
	"context"
	"fmt"
	"io"
	"time"
)

func ExamplePinger() {

	// 创建一个可以被取消的contest对象，cancel函数被用来在适当的时候结束pinger goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// 使用io.Pipe 创建一对读写管道
	r, w := io.Pipe()
	// 用于在Pinger goroutine结束时发送信号。
	done := make(chan struct{})
	// 创建了一个resetTimer通道，它是一个带缓冲的通道，用于发送新的ping间隔到Pinger函数。
	resetTimer := make(chan time.Duration, 1)
	// 向resetTimer通道发送了一个时间间隔，这是首次ping的间隔
	resetTimer <- time.Second // initial ping interval

	// 在新的goroutine中启动Ping函数，并在函数结束时关闭done 通道
	go func() {
		Pinger(ctx, w, resetTimer)
		close(done)
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d
		}
	}
	now := time.Now()
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("received %q (%s)\n", buf[:n],
		time.Since(now).Round(100*time.Millisecond))
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}
	cancel()
	<-done //ensures the pinger exits after canceling the context

}
