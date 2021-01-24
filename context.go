package gee

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Knight-7/gee/binding"
)

var (
	jsonContentType = []string{"application/json; charset=utf-8"}
	textContentType = []string{"text/plain; charset=utf-8"}
	xmlContentType  = []string{"application/xml; charset=utf-8"}
	yamlContentType = []string{"application/x-yaml; charset=utf-8"}
)

type H map[string]interface{}

type Context struct {
	// original obj
	Writer http.ResponseWriter
	Req    *http.Request

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
	c.Writer = w
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
	c.Abort()
	c.JSON(code, err)
}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
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
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) writeContentType(val []string) {
	c.SetHeader("Content-Type", val[0])
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.writeContentType(textContentType)
	c.Status(code)
	_, _ = c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.writeContentType(jsonContentType)
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) XML(code int, obj interface{}) {
	c.writeContentType(xmlContentType)
	c.Status(code)
	encoder := xml.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) YAML(code int, obj interface{}) {
	c.writeContentType(yamlContentType)
	c.Status(code)
	encoder := yaml.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, _ = c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

// TODO: 学习 Golang 中 cookie 的知识和使用方法
func (c *Context) SetCookie(name string, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}

	http.SetCookie(c.Writer, &http.Cookie{
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

func (c *Context) ContentType() string {
	return c.Req.Header.Get("Content-Type")
}

func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Method, c.ContentType())
	return c.MustBindWith(obj, b)
}

func (c *Context) BindJSON(obj interface{}) error {
	return c.MustBindWith(obj, binding.JSON)
}

func (c *Context) BindXML(obj interface{}) error {
	return c.MustBindWith(obj, binding.XML)
}

func (c *Context) BindYAML(obj interface{}) error {
	return c.MustBindWith(obj, binding.YAML)
}

func (c *Context) MustBindWith(obj interface{}, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.Req, obj)
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
