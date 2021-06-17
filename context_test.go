package gee

import (
	"fmt"
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

func TestRendering(t *testing.T) {
	engine := Default()
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/someJSON/:name", func(c *Context) {
			name := c.Param("name")
			c.JSON(http.StatusOK, H{
				"name":         name,
				"Content-Type": c.Writer.Header().Get("Content-Type"),
			})
		})
		v1.GET("/someXML/:name", func(c *Context) {
			name := c.Param("name")
			c.XML(http.StatusOK, H{
				"name":         name,
				"Content-type": c.Writer.Header().Get("Content-Type"),
			})
		})
		v1.GET("/someYAML/:name", func(c *Context) {
			name := c.Param("name")
			c.YAML(http.StatusOK, H{
				"name":         name,
				"Content-Type": c.Writer.Header().Get("Content-Type"),
			})
		})
		v1.GET("/hello/:name", func(c *Context) {
			name := c.Param("name")
			c.String(http.StatusOK, "Hello, %s.\n", name)
		})
		v1.GET("/redirect/:name", func(c *Context) {
			name := c.Param("name")
			c.Redirect(http.StatusTemporaryRedirect,  fmt.Sprintf("/api/v1/hello/%s", name))
		})
	}
	engine.Run(":2020")
}

// 写法1
func TestContext_HTML(t *testing.T) {
	engine := Default()
	engine.LoadHTMLGlob("testdata/static/*")
	engine.GET("/index/:name", func(c *Context) {
		c.HTML(http.StatusOK, "index.html", H{
			"title": "index title",
			"body":  fmt.Sprintf("Welcome %s", c.Param("name")),
		})
	})
	engine.Run(":2020")
}

// 写法2
func TestContext_HTML2(t *testing.T) {
	engine := Default()
	engine.LoadHTMLGlob("testdata/**/*")
	engine.GET("/index", func(c *Context) {
		c.HTML(http.StatusOK, "static/hello.html", H{
			"title": "html title",
			"name":  "Knight",
		})
	})
	engine.Run(":2020")
}

func TestContext_File(t *testing.T) {
	engine := Default()
	v1 := engine.Group("/api/v1")
	{
		// 文件上传
		v1.POST("/upload", func(c *Context) {
			fileHeader, err := c.FormFile("file")
			if err != nil {
				c.Fail(http.StatusInternalServerError, err.Error())
				return
			}
			err = c.SaveFile(fileHeader, fileHeader.Filename)
			if err != nil {
				c.Fail(http.StatusInternalServerError, err.Error())
				return
			}
			c.JSON(http.StatusOK, nil)
		})

		// 文件下载
		v1.GET("/download/*filename", func(c *Context) {
			filename := c.Param("filename")
			filename = fmt.Sprintf("testdata/static/%s", filename)
			fmt.Println(filename)
			c.File(filename)
		})
	}
	engine.Run(":2020")
}
