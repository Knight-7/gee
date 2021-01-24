package binding

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
)

type yamlBinding struct {}

func (b yamlBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return decodeYAML(req, obj)
}

func decodeYAML(req *http.Request, obj interface{}) error {
	decoder := yaml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}
