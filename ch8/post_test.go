package ch8

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type User struct {
	First string
	Last  string
}

func handlePostUser(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(r io.ReadCloser) {
			_, _ = io.Copy(io.Discard, r)
			_ = r.Close()
		}(r.Body)

		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			t.Error(err)
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func TestPostUser(t *testing.T) {
	// 新建一个测试服务器，传入一个处理方法
	ts := httptest.NewServer(http.HandlerFunc(handlePostUser(t)))
	defer ts.Close()

	// 向测试服务器发送GET请求
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	//
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusMethodNotAllowed, resp.StatusCode)
	}

	// 字节缓冲区
	buf := new(bytes.Buffer)
	u := User{First: "Adam", Last: "Woodbeck"}
	// 使用JSON编码器将结构体 u 编码并写入字节缓冲区 buf 中，以准备将其作为请求体发送到服务器。
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	//
	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("excepted status %d; actual status %d",
			http.StatusAccepted, resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestMultipartPost(t *testing.T) {
	// 设置缓冲区
	reqBody := new(bytes.Buffer)
	// 创建multipart写入器，指定写入的对象
	w := multipart.NewWriter(reqBody)

	for k, v := range map[string]string{
		"data":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		filePath, err := w.CreateFormFile(fmt.Sprintf("files%d", i+1),
			filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(filePath, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://httpbin.org/post", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
	t.Logf("\n%s", b)
}
