package ch3

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	// 定义一个时间间隔 变量
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			// 当一个time.NewTimer类型的定时器到期后，会发送当前时间
			// 到C 这个通道。timer.Stop这个函数的行为是停止一个定时器
			// 如果定时器已经到期，则返回false。这也意味着定时器
			// 向C发送数据这一事件已经发生，这时我们通常需要从C通道中取出
			// 一个值，以确保C通道被清空。
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
