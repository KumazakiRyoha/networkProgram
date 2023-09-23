package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
*
一旦Write方法被调用，如果此时状态码还没有被设置，那么它会默认为200 OK。一旦状态码被设置，
无论是明确地还是隐式地，都不能再改变它。因此，当需要返回非200 OK的状态码时，应该先调用WriteHeader方法，
然后再写响应体。
*/
func TestHandlerWriteHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("BadRequest"))
		w.WriteHeader(http.StatusBadRequest)
	}
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()
	handler(w, r)
	t.Logf("Response status: %q", w.Result().Status)

	handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}
	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()
	handler(w, r)
	t.Logf("Response status: %q", w.Result().Status)
}
