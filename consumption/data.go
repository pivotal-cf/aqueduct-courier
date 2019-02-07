package consumption

import (
	"io"
)

type Data struct {
	reader   io.Reader
	dataType string
}

func NewData(reader io.Reader, dataType string) Data {
	return Data{reader: reader, dataType: dataType}
}

func (d Data) Name() string {
	return d.dataType
}

func (d Data) Content() io.Reader {
	return d.reader
}

func (d Data) MimeType() string {
	return "application/json"
}

func (d Data) Type() string {
	return ""
}

func (d Data) DataType() string {
	return d.dataType
}
