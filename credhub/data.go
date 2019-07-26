package credhub

import (
	"fmt"
	"io"

	"github.com/pivotal-cf/telemetry-utils/data"
)

type Data struct {
	reader io.Reader
}

func NewData(reader io.Reader) Data {
	return Data{reader: reader}
}

func (d Data) Name() string {
	return fmt.Sprintf("%s_%s", d.Type(), d.DataType())
}

func (d Data) Content() io.Reader {
	return d.reader
}

func (d Data) MimeType() string {
	return "application/json"
}

func (d Data) Type() string {
	return data.DirectorProductType
}

func (d Data) DataType() string {
	return data.CertificatesDataType
}
