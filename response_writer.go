package gee

import (
	"bufio"
	"net"
	"net/http"
)

// 自定义一个 ResponseWriter 接口，来保存和获取 status 状态
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	http.CloseNotifier

	Status() int
}

type response struct {
	http.ResponseWriter
	status int
}

func newResponse(writer http.ResponseWriter) *response {
	return &response{
		ResponseWriter: writer,
		status:         http.StatusOK,
	}
}

func (w *response) Status() int {
	return w.status
}

// 覆盖 WriteHeader 方法，这样其他地方调用时会调用此方法，并将 status 保存到其中
func (w *response) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *response) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *response) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
