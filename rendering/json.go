package rendering

import (
	"encoding/json"
	"net/http"
)

type JSON struct {
	Data interface{}
}

func (r JSON) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	data, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
