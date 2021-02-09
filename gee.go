package gee

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

type HandlerFunc func(*Context)

type Engine struct {
	// 路由
	*RouterGroup
	router        *router
	groups        []*RouterGroup
	// HTML 渲染
	htmlTemplates *template.Template
	funcMap       template.FuncMap
	// Context 池（减少 GC 带来的消耗）
	pool          sync.Pool
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) allocateContext() *Context {
	return &Context{engine: engine}
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := engine.pool.Get().(*Context)
	c.reset(w, r)
	c.handlers = middlewares
	engine.router.handle(c)

	engine.pool.Put(c)
}

// Graceful shutdown server
func (engine *Engine) Run(addr string) {
	assert1(addr != "", "Server address can't be null")

	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		fmt.Printf("Server listen at %s successfully. Use 'Ctrl + C' to stop Server\n", addr)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Printf("Shutdown serve...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown faild: %s\n", err)
	}
	log.Printf("Serve exiting")
}
