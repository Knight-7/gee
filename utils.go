package gee

import "encoding/xml"

type H map[string]interface{}

func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for k, v := range h {
		startElement := xml.StartElement{
			Name: xml.Name{Space: "", Local: k},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(v, startElement); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func assert1(flag bool, text string) {
	if !flag {
		panic(text)
	}
}
