package ch3

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	// 从reset通道中拿到一个时间间隔
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}
	// 创建一个定时器
	timer := time.NewTimer(interval)
	// 函数结执行时清空`C`通道
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	// 无线循环，等待新的时间间隔或者写出数据
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}
		_ = timer.Reset(interval)
	}

}
