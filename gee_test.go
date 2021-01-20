package gee

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupEngine(t *testing.T) *Engine {
	engine := Default()
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/user/:id", func(c *Context) {
			id := c.Param("id")
			c.String(http.StatusOK, "Hello %s\n", id)
		})
		v1.POST("/user", func(c *Context) {
			c.JSON(http.StatusOK, H{
				"username": c.PostForm("user"),
				"password": c.PostForm("pass"),
			})
		})
		v1.PUT("/user/:id/update", func(c *Context) {
			id := c.Param("id")
			c.String(http.StatusOK, "%s update ok\n", id)
		})
		v1.DELETE("/user/:id/delete", func(c *Context) {
			id := c.Param("id")
			c.String(http.StatusOK, "%s delete ok\n", id)
		})
	}
	v2 := engine.Group("/api/v2")
	v2.Use(func(c *Context) {
		start := time.Now()
		c.Fail(http.StatusInternalServerError, "internal server error")
		log.Printf("[%d] %s in %v\n", c.StatusCode, c.Req.RequestURI, time.Since(start))
	})
	{
		v2.GET("/hello", func(c *Context) {
			name := c.Query("name")
			c.String(http.StatusOK, "hello %s", name)
		})
	}

	return engine
}

func TestGee(t *testing.T) {
	engine := setupEngine(t)

	getWriter := httptest.NewRecorder()
	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/user/34", nil)
	engine.ServeHTTP(getWriter, getReq)
	assert.Equal(t, http.StatusOK, getWriter.Code)
	assert.Equal(t, "Hello 34\n", getWriter.Body.String())

	postWriter := httptest.NewRecorder()
	postReq, _ := http.NewRequest(http.MethodPost, "/api/v1/user", nil)
	engine.ServeHTTP(postWriter, postReq)
	assert.Equal(t, http.StatusOK, postWriter.Code)
	t.Log(postWriter.Body.String())

	putWriter := httptest.NewRecorder()
	putReq, _ := http.NewRequest(http.MethodPut, "/api/v1/user/34/update", nil)
	engine.ServeHTTP(putWriter, putReq)
	assert.Equal(t, http.StatusOK, putWriter.Code)
	assert.Equal(t, "34 update ok\n", putWriter.Body.String())

	deleteWriter := httptest.NewRecorder()
	deleteReq, _:= http.NewRequest(http.MethodDelete, "/api/v1/user/34/delete", nil)
	engine.ServeHTTP(deleteWriter, deleteReq)
	assert.Equal(t, http.StatusOK, deleteWriter.Code)
	assert.Equal(t, "34 delete ok\n", deleteWriter.Body.String())

	defer func() {
		r := recover()
		assert.Equal(t, fmt.Sprintf("The new path '/api/v2/knight' is conflict with path '/api/v2/:name'"), r)
	}()

	engine.POST("/api/v2/:name", nil)
	engine.POST("/api/v2/knight", nil)
}

func TestRouterConflict(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "The new path '/api/v1/user/hello/:name' is conflict with path '/api/v1/user/hello/knight'", r)
	}()

	engine := Default()
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/user/hello/knight", func(c *Context) {
			c.String(http.StatusOK, c.Path)
		})
		v1.GET("/user/hello/:name/ecs/:id", func(c *Context) {
			c.String(http.StatusOK, c.Path)
		})
		v1.GET("/user/hello/:name/ecs/:id/update", func(c *Context) {
			c.String(http.StatusOK, c.Path)
		})
		v1.GET("/user/hello/:name", func(c *Context) {
			c.String(http.StatusOK, c.Path)
		})
	}

	put1Write := httptest.NewRecorder()
	put1Req, _ := http.NewRequest(http.MethodGet, "/api/v1/user/hello/knight", nil)
	engine.ServeHTTP(put1Write, put1Req)
	put2Write := httptest.NewRecorder()
	put2Req, _ := http.NewRequest(http.MethodGet, "/api/v1/user/hello/kobe", nil)
	engine.ServeHTTP(put2Write, put2Req)
	t.Log(put1Write.Body.String())
	t.Log(put2Write.Body.String())
}

func TestAddRouter1(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "Path must begin with '/'", r)
	}()

	engine := Default()
	engine.GET("api/user", nil)
}

func TestAddRouter2(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "Handler can not be nil", r)
	}()

	engine := Default()
	engine.GET("/api/v1/user", nil)
}

func TestAddRouter3(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "HTTP method can not be empty", r)
	}()

	engine := Default()
	engine.addRouter("", "/api/v1", nil)
}

func TestEngineRun(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, "Server address can't be null", r)
	}()

	engine := Default()
	engine.Run("")
}
