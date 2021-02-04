package rendering

import "net/http"

type Data struct {
	ContentType string
	Data        []byte
}

func (r Data) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	_, err := w.Write(r.Data)
	if err != nil {
		return err
	}
	return nil
}

func (r Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
}
