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

	/*encoder := xml.NewEncoder(w)
	if err := encoder.Encode(r.Data); err != nil {
		return err
	}

	return nil*/
	return xml.NewEncoder(w).Encode(r.Data)
}

func (r XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, xmlContentType)
}
