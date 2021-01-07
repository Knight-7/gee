package gee

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func logger(c *Context) {
	start := time.Now()

	c.Next()

	log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(start))
}

func TestGee(t *testing.T) {
	engine := New()
	engine.Use(logger)

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
			c.String(http.StatusOK, "%s update ok", id)
		})
		v1.DELETE("/user/:id/delete", func(c *Context) {
			id := c.Param("id")
			c.String(http.StatusOK, "%s delete ok", id)
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

	if err := engine.Run("localhost:3434"); err != nil {
		assert.NoError(t, err)
	}
}
