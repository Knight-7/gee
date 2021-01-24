package binding

import "net/http"

const (
	MIMEJSON              = "application/json"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEYAML              = "application/x-yaml"
)

var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	YAML          = yamlBinding{}
	Form          = formBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = multipartBinding{}
)

type Binding interface {
	Bind(*http.Request, interface{}) error
}

func Default(method, contentType string) Binding {
	// 当请求的 Method 是时 GET 时，此时解析的是 URL 上的参数
	if method == http.MethodGet {
		return Form
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	case MIMEYAML:
		return YAML
	case MIMEPlain:
		return Form
	case MIMEPOSTForm:
		return FormPost
	case MIMEMultipartPOSTForm:
		return FormMultipart
	default:
		return Form
	}
}
