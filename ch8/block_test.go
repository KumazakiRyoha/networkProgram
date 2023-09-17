package ch8

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	_, _ = http.Get(ts.URL)
	t.Fatal("client did not indefinitely block")
}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	// 创建并启动一个新的 HTTP 测试服务器.这个处理器定义了当 HTTP
	// 请求发送到服务器时，服务器应该如何响应。通常，这是一个模拟的响应，用于测试客户端的行为。
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	_ = resp.Body.Close()

}
