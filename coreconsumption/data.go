package coreconsumption

import (
	"io"
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
	return d.dataType
}

func (d Data) Content() io.Reader {
	return d.reader
}

func (d Data) MimeType() string {
	return "application/json"
}

func (d Data) Type() string {
	return d.productType
}

func (d Data) DataType() string {
	return d.dataType
}
