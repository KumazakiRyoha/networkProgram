package ch3

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	// 设置一个超时时间，时间为当前时间的五秒后
	dl := time.Now().Add(5 * time.Second)
	// 创建一个带有截至日期的context。如果到达戒截至日期，或者调用了返回
	// 的cancel函数，context函数会被取消。
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()
	var d net.Dialer

	d.Control = func(_, address string, _ syscall.RawConn) error {
		// Sleep long enough to reach the context's deadline
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}

	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err == nil {
		conn.Close()
		t.Fatal("cconnection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}

}
