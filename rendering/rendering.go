package rendering

import (
	"net/http"
)

var (
	jsonContentType = []string{"application/json; charset=utf-8"}
	textContentType = []string{"text/plain; charset=utf-8"}
	xmlContentType  = []string{"application/xml; charset=utf-8"}
	yamlContentType = []string{"application/x-yaml; charset=utf-8"}
	htmlContentType = []string{"text/html; charset=utf-8"}
)

type Render interface {
	Render(http.ResponseWriter) error
	WriteContentType(http.ResponseWriter)
}

func writeContentType(w http.ResponseWriter, val []string) {
	header := w.Header()
	if v := header["Content-Type"]; len(v) == 0 {
		header["Content-Type"] = val
	}
}
