package ch5

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestListenPacketUDP(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	// 创建一个服务器，它在本地地址127.0.0.1的任意可用端口上监听。
	serverUDP, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	// 创建了一个UDP客户端，它也在本地地址的任意可用端口上监听。
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = client.Close() }()

	// 创建另一个UDP客户端（称为interloper）。这个客户端的目的是向上面创建的client发送一个"中断"消息。
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// interloper向client发送一个"pardon me"的消息。
	interrupt := []byte("pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()
	// 验证消息是否成功发送
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	// 主客户端client向服务器发送一个"ping"消息。
	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverUDP)
	if err != nil {
		t.Fatal(err)
	}

	// client尝试从其UDP套接字读取响应。
	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// 接收到的第一个消息
	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("excepted reply %q;actual reply %q", interrupt, buf[:n])
	}
	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf("expected message from %q; actual sender is %q",
			interloper.LocalAddr(), addr)
	}

	// 验证接收到的第二个消息
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}
	if addr.String() != serverUDP.String() {
		t.Errorf("expected message from %q; actual sender is %q",
			serverUDP, addr)
	}
}
