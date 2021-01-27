package rendering

import (
	"gopkg.in/yaml.v3"
	"net/http"
)

type YAML struct {
	Data interface{}
}

func (r YAML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	data, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (r YAML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, yamlContentType)
}
