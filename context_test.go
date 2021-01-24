package gee

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setUpContext(engine *Engine, w http.ResponseWriter, r *http.Request) *Context {
	c := engine.allocateContext()
	c.reset(w, r)
	return c
}

func TestSetAndGet(t *testing.T) {
	engine := setupEngine(t)
	writer := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/hello/:name", nil)
	c := setUpContext(engine, writer, req)

	c.Set("session", "session_val")
	c.Set("sjdlfdjsf", "jflsjfjsd")

	val1, ok1 := c.Get("session")
	assert.True(t, ok1)
	assert.Equal(t, val1, "session_val")

	val2, ok2 := c.Get("sjdlfdjsf")
	assert.True(t, ok2)
	assert.Equal(t, val2, "jflsjfjsd")

	val3, ok3 := c.Get("sfsfjsf")
	assert.False(t, ok3)
	assert.Equal(t, val3, nil)

	val4 := c.MustGet("session")
	assert.Equal(t, val4, "session_val")

	defer func() {
		r := recover()
		assert.Equal(t, "Key \"sdflksjf\" not exists", r)
	}()

	_ = c.MustGet("sdflksjf")
}

func TestBind(t *testing.T) {
	engine := setupEngine(t)
	engine.Run("localhost:2020")
}
