package rendering

import (
	"fmt"
	"net/http"
)

type Redirect struct {
	Code int
	Location string
	Request *http.Request
}

func (r Redirect) Render(w http.ResponseWriter) error {
	if r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect {
		panic(fmt.Sprintf("Can not redirect with code %d", r.Code))
	}
	http.Redirect(w, r.Request, r.Location, r.Code)
	return nil
}

func (r Redirect) WriteContentType(w http.ResponseWriter) {}
