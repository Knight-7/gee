package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()

		c.Next()

		log.Printf("[%d] %-7s %s in %v\n", c.Writer.Status(), c.Method, c.Request.RequestURI, time.Since(start))
	}
}
