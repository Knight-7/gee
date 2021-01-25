package rendering

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data interface{}
}

func (r XML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	data, err := xml.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (r XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, xmlContentType)
}
