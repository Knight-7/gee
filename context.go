package gee

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Knight-7/gee/binding"
	"github.com/Knight-7/gee/rendering"
)

const (
	defaultMaxMemory = 32 << 20
)

// Context 请求的上下文
type Context struct {
	Writer      ResponseWriter
	Request     *http.Request
	Path        string
	Method      string
	Params      map[string]string
	middlewares []HandlerFunc // record middlewares
	index       int
	engine      *Engine // engine pointer
	Keys        map[string]interface{}
	mu          sync.RWMutex
	sameSite    http.SameSite // cookie
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Writer = newResponse(w)
	c.Request = r
	c.Path = c.Request.URL.Path
	c.Method = c.Request.Method
	c.Params = nil
	c.middlewares = nil
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
	for ; c.index < len(c.middlewares); c.index++ {
		c.middlewares[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.Abort()
	c.JSON(code, err)
}

func (c *Context) Abort() {
	c.index = len(c.middlewares)
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Status(code int) {
	if code > 0 {
		c.Writer.WriteHeader(code)
	}
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, rendering.String{Format: format, Value: values})
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, rendering.JSON{Data: obj})
}

func (c *Context) XML(code int, obj interface{}) {
	c.Render(code, rendering.XML{Data: obj})
}

func (c *Context) YAML(code int, obj interface{}) {
	c.Render(code, rendering.YAML{Data: obj})
}

func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, rendering.Data{ContentType: contentType, Data: data})
}

func (c *Context) Redirect(code int, location string) {
	c.Render(-1, rendering.Redirect{
		Code:     code,
		Location: location,
		Request:  c.Request,
	})
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.Render(code, rendering.HTML{
		Name:     name,
		Data:     data,
		Template: c.engine.htmlTemplates,
	})
}

func (c *Context) Render(code int, r rendering.Render) {
	if !c.bodyCanWriteContentWithStatus(code) {
		r.WriteContentType(c.Writer)
		return
	}

	r.WriteContentType(c.Writer)
	c.Status(code)

	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}

func (c *Context) bodyCanWriteContentWithStatus(code int) bool {
	switch {
	// 状态码 1** 表示服务器收到消息，需要请求者继续操作
	case code >= 100 && code <= 199:
		return false
	// 状态码 204 无内容，服务器处理成功，但没有返回内容
	case code == http.StatusNoContent:
		return false
	// 状态码 304 所请求的资源为修改，服务器只返回状态码，不返回任何资源
	case code == http.StatusNotModified:
		return false
	}
	return true
}

func (c *Context) FormFile(filename string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, err
		}
	}

	file, fh, err := c.Request.FormFile(filename)
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	return fh, err
}

func (c *Context) SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

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
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}

	val, _ := url.QueryUnescape(cookie.Value)

	return val, nil
}

func (c *Context) SetSameSite(sameSite http.SameSite) {
	c.sameSite = sameSite
}

func (c *Context) ContentType() string {
	return filterContent(c.Request.Header.Get("Content-Type"))
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

func (c *Context) BindURL(obj interface{}) error {
	return c.MustBindWith(obj, binding.Form)
}

func (c *Context) MustBindWith(obj interface{}, b binding.Binder) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBindWith(obj interface{}, b binding.Binder) error {
	return b.Bind(c.Request, obj)
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
