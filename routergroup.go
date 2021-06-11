package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
)

type Routers interface {
	Group(string) Routers
	Router
}

type Router interface {
	Use(...HandlerFunc)

	GET(string, HandlerFunc)
	Header(string, HandlerFunc)
	POST(string, HandlerFunc)
	PUT(string, HandlerFunc)
	DELETE(string, HandlerFunc)
	Connect(string, HandlerFunc)
	Options(string, HandlerFunc)
	Trace(string, HandlerFunc)
	Patch(string, HandlerFunc)
	Any(string, HandlerFunc)

	Static(string, string)
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func (group *RouterGroup) Group(prefix string) Routers {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRouter(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("%-7s - %s\n", method, pattern)
	group.engine.router.addRouter(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodGet, pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodPost, pattern, handler)
}

func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodPut, pattern, handler)
}

func (group *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodDelete, pattern, handler)
}

func (group *RouterGroup) Header(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodHead, pattern, handler)
}

func (group *RouterGroup) Connect(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodConnect, pattern, handler)
}

func (group *RouterGroup) Options(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodOptions, pattern, handler)
}

func (group *RouterGroup) Trace(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodTrace, pattern, handler)
}

func (group *RouterGroup) Patch(pattern string, handler HandlerFunc) {
	group.addRouter(http.MethodPatch, pattern, handler)
}

func (group *RouterGroup) Any(pattern string, handler HandlerFunc) {
	group.GET(pattern, handler)
	group.Header(pattern, handler)
	group.POST(pattern, handler)
	group.PUT(pattern, handler)
	group.DELETE(pattern, handler)
	group.Connect(pattern, handler)
	group.Options(pattern, handler)
	group.Trace(pattern, handler)
	group.Patch(pattern, handler)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}
