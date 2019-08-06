package credhub_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"
)

var _ = Describe("Data", func() {

	It("returns a name", func() {
		d := NewData(
			strings.NewReader(""),
		)
		Expect(d.Name()).To(Equal(collector_tar.DirectorProductType + "_" + collector_tar.CertificatesDataType))
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
		Expect(d.Type()).To(Equal(collector_tar.DirectorProductType))
	})

	It("returns the data type", func() {
		d := NewData(nil)
		Expect(d.DataType()).To(Equal(collector_tar.CertificatesDataType))
	})

})
