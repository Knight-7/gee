package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var jsonContentType = []string{"application/json; charset=utf-8"}
var textContentType = []string{"text/plain"}

type H map[string]interface{}

type Context struct {
	// original obj
	Write http.ResponseWriter
	Req   *http.Request

	// request info
	Path   string
	Method string
	Params map[string]string

	// response info
	StatusCode int

	// middleware
	handlers []HandlerFunc
	index    int

	// engine pointer
	engine *Engine

	// context
	Keys map[string]interface{}
	mu   sync.RWMutex

	// cookie
	sameSite http.SameSite
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Write = w
	c.Req = r
	c.Path = c.Req.URL.Path
	c.Method = c.Req.Method
	c.Params = nil
	c.StatusCode = 0
	c.handlers = nil
	c.index = -1
	c.Keys = nil
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.Keys[key]
	return value, ok
}

func (c *Context) MustGet(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.Keys[key]
	if !ok {
		panic("Key \"" + key + "\" not exists")
	}

	return value
}

func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, err)
}

func (c *Context) PostForm(key string) string {
	return c.Req.PostFormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Write.WriteHeader(code)
}

func (c *Context) SetHeader(key, value string) {
	c.Write.Header().Set(key, value)
}

func (c *Context) writeContentType(key string, val []string) {
	c.SetHeader(key, val[0])
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.writeContentType("Content-Type", textContentType)
	c.Status(code)
	_, _ = c.Write.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.writeContentType("Content-Type", jsonContentType)
	c.Status(code)
	encoder := json.NewEncoder(c.Write)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Write, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, _ = c.Write.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Write, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

func (c *Context) SetCookie(name string, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}

	http.SetCookie(c.Write, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Req.Cookie(name)
	if err != nil {
		return "", err
	}

	val, _ := url.QueryUnescape(cookie.Value)

	return val, nil
}

// set with cookie
func (c *Context) SetSameSite(sameSite http.SameSite) {
	c.sameSite = sameSite
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Value(key interface{}) interface{} {
	return nil
}
