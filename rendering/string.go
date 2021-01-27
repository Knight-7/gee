package rendering

import (
	"fmt"
	"net/http"
)

type String struct {
	Format string
	Value  []interface{}
}

func (r String) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	formatData := fmt.Sprintf(r.Format, r.Value...)
	_, err := w.Write([]byte(formatData))
	if err != nil {
		return err
	}
	return nil
}

func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, textContentType)
}
