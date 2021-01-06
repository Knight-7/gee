package gee

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRouter(http.MethodGet, "/", nil)
	r.addRouter(http.MethodGet, "/hello/:name", nil)
	r.addRouter(http.MethodGet, "/static/*filename", nil)
	r.addRouter(http.MethodGet, "/hello/user/:id", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	assert.Equal(t, parsePattern("/hello/a"), []string{"hello", "a"})

	assert.Equal(t, parsePattern("/hello/:name"), []string{"hello", ":name"})

	assert.Equal(t, parsePattern("/static/*filename"), []string{"static", "*filename"})
}

func TestGetRouter(t *testing.T) {
	r := newTestRouter()

	n1, params := r.getRouter(http.MethodGet, "/hello/a")
	assert.NotNil(t, n1)
	assert.Equal(t, map[string]string{"name": "a"}, params)

	n2, params := r.getRouter(http.MethodGet, "/hello/knight")
	assert.NotNil(t, n2)
	assert.Equal(t, "/hello/:name", n2.pattern)
	assert.Equal(t, map[string]string{"name": "knight"}, params)

	n3, params := r.getRouter(http.MethodGet, "/")
	assert.NotNil(t, n3)
	assert.Equal(t, "/", n3.pattern)
	assert.Equal(t, make(map[string]string), params)

	n4, params := r.getRouter(http.MethodGet, "/static/tmp/tmp.css")
	assert.NotNil(t, n4)
	assert.Equal(t, "/static/*filename", n4.pattern)
	assert.Equal(t, map[string]string{"filename": "tmp/tmp.css"}, params)

	n5, params := r.getRouter(http.MethodGet, "/hello/user/7")
	assert.NotNil(t, n5)
	assert.Equal(t, "/hello/user/:id", n5.pattern)
	assert.Equal(t, map[string]string{"id": "7"}, params)
}
