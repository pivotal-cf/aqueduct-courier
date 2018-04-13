package opsmanager

import (
	"fmt"
	"io"
)

const (
	JSONDataType    = "json"
	DefaultDataType = JSONDataType
)

type Data struct {
	reader io.Reader
	name   string
	kind   string
}

func NewData(reader io.Reader, name, kind string) Data {
	return Data{reader: reader, name: name, kind: kind}
}

func (d Data) Name() string {
	return fmt.Sprintf("%s_%s", d.name, d.kind)
}

func (d Data) Content() io.Reader {
	return d.reader
}

func (d Data) ContentType() string {
	return DefaultDataType
}
