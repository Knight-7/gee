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
	// Request 请求和 ResponseWriter
	Writer  ResponseWriter
	Request *http.Request

	// request info
	Path   string
	Method string
	Params map[string]string

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
	c.Writer = newResponse(w)
	c.Request = r
	c.Path = c.Request.URL.Path
	c.Method = c.Request.Method
	c.Params = nil
	c.handlers = nil
	c.index = -1
	c.Keys = nil
}

// 在 Context 中设置数据
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// 获取设置在 Context 中的数据
func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.Keys[key]
	return value, ok
}

// 获取设置在 Context 中的数据，不存在将 panic
func (c *Context) MustGet(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.Keys[key]
	if !ok {
		panic("Key \"" + key + "\" not exists")
	}

	return value
}

// 调用下一个中间件
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

// 停止响应
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// 直接设置状态码并停止响应
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// 请求为 POST 时，获取 key 对应的数据
func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

// 获取请求 URL 中的
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// 获取请求 URL 中的参数
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// 设置响应的状态码
func (c *Context) Status(code int) {
	if code > 0 {
		c.Writer.WriteHeader(code)
	}
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// 响应字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, rendering.String{Format: format, Value: values})
}

// 响应 JSON 数据
func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, rendering.JSON{Data: obj})
}

// 响应 XML 数据
func (c *Context) XML(code int, obj interface{}) {
	c.Render(code, rendering.XML{Data: obj})
}

// 响应 YAML 数据
func (c *Context) YAML(code int, obj interface{}) {
	c.Render(code, rendering.YAML{Data: obj})
}

// 响应数据
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, rendering.Data{ContentType: contentType, Data: data})
}

// 重定向
func (c *Context) Redirect(code int, location string) {
	c.Render(-1, rendering.Redirect{
		Code:     code,
		Location: location,
		Request:  c.Request,
	})
}

// 响应 HTML 格式
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
	case  code == http.StatusNotModified:
		return false
	}
	return true
}

// 该方法解析 multipart/form-data 中上传的文件，并返回 multipart.FileHeader
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

// 保存文件
// 参数 file 是通过 FormFile 获取的文件， dst 是保存的路径
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

// 响应返回文件
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
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
	cookie, err := c.Request.Cookie(name)
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

// 返回 Request 中的 Content-Type
func (c *Context) ContentType() string {
	return filterContent(c.Request.Header.Get("Content-Type"))
}

// 数据绑定，支持绑定 URL PostForm MultipartForm JSON XML YAML 的数据
func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Method, c.ContentType())
	return c.MustBindWith(obj, b)
}

// 绑定 Content-Type 为 application/json 的数据
func (c *Context) BindJSON(obj interface{}) error {
	return c.MustBindWith(obj, binding.JSON)
}

// 绑定 Content-Type 为 application/xml 的数据
func (c *Context) BindXML(obj interface{}) error {
	return c.MustBindWith(obj, binding.XML)
}

// 绑定 Content-Type 为 application/x-yaml 的数据
func (c *Context) BindYAML(obj interface{}) error {
	return c.MustBindWith(obj, binding.YAML)
}

// 当 Method 为 Get 是，绑定 URL 上的参数
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
