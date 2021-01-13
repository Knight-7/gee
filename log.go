package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()

		c.Next()

		log.Printf("[%d] %s %s in %v\n", c.StatusCode, c.Method, c.Req.RequestURI, time.Since(start))
	}
}
