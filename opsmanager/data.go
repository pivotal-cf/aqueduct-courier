package opsmanager

import (
	"fmt"
	"io"
)

const (
	JSONDataType    = "application/json"
	DefaultDataType = JSONDataType
)

type Data struct {
	reader      io.Reader
	productType string
	dataType    string
}

func NewData(reader io.Reader, productType, dataType string) Data {
	return Data{reader: reader, productType: productType, dataType: dataType}
}

func (d Data) Name() string {
	return fmt.Sprintf("%s_%s", d.productType, d.dataType)
}

func (d Data) Content() io.Reader {
	return d.reader
}

func (d Data) MimeType() string {
	return DefaultDataType
}

func (d Data) Type() string {
	return d.productType
}
