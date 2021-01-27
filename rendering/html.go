package rendering

import (
	"html/template"
	"net/http"
)

type HTML struct {
	Name   string
	Data   interface{}
	Template *template.Template
}

func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	err := r.Template.ExecuteTemplate(w, r.Name, r.Data)
	if err != nil {
		return err
	}
	return nil
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
