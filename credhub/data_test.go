package credhub_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/telemetry-utils/data"
)

var _ = Describe("Data", func() {

	It("returns a name", func() {
		d := NewData(
			strings.NewReader(""),
		)
		Expect(d.Name()).To(Equal(data.DirectorProductType + "_" + data.CertificatesDataType))
	})

	It("returns content for the data", func() {
		dataReader := strings.NewReader("best-data")
		d := NewData(dataReader)
		Expect(d.Content()).To(Equal(dataReader))
	})

	It("returns json as data type", func() {
		d := NewData(nil)
		Expect(d.MimeType()).To(Equal("application/json"))
	})

	It("returns the product type", func() {
		d := NewData(nil)
		Expect(d.Type()).To(Equal(data.DirectorProductType))
	})

	It("returns the data type", func() {
		d := NewData(nil)
		Expect(d.DataType()).To(Equal(data.CertificatesDataType))
	})

})
