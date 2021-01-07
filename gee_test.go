package gee

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGee(t *testing.T) {
	e := New()
	e.GET("/user/:id", func(c *Context) {
		id := c.Param("id")
		c.HTML(http.StatusOK, fmt.Sprintf("<h1> Hello, %s </h1>", id))
	})

	e.POST("/user", func(c *Context) {
		user := c.PostForm("user")
		c.JSON(http.StatusOK, H{
			"user": user,
			"form": c.Req.PostForm,
		})
	})

	e.PUT("/user/:id/update", func(c *Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "%s update ok", id)
	})

	e.DELETE("/user/:id/delete", func(c *Context) {
		id := c.Param("id")
		c.Data(http.StatusOK, []byte(id))
	})

	e.Run("localhost:3434")
}
