package handlers

import (
	"html/template"
	"io"
	"net/http"
)

// 创建并解析了一个简单的模板，并确保没有解析错误。创建的模板可以用于将传递的数据渲染为 "Hello,数据!" 的形式。
var t = template.Must(template.New("hello").Parse("Hello,{{.}}!"))

func DefaultHandler() http.Handler {
	// http.HandlerFunc(匿名函数) <- 将匿名函数转换为HandlerFunc 类型并返回
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func(r io.ReadCloser) {
				io.Copy(io.Discard, r)
			}(r.Body)

			var b []byte
			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				// 读取请求体中的所有信息，并作为响应内容
				b, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error",
						http.StatusInternalServerError)
					return
				}
			default:
				// not RFC-compile due to locak of "Allow" hearder
				http.Error(w, "Method not allowed",
					http.StatusMethodNotAllowed)
				return
			}
			t.Execute(w, string(b))
		},
	)
}
