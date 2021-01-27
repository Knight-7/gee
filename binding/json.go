package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type jsonBinding struct {}

func (b jsonBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return decodeJson(req, obj)
}

func decodeJson(req *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}
