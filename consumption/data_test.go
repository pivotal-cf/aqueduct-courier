package consumption_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/consumption"
	"github.com/pivotal-cf/aqueduct-utils/data"
)

var _ = Describe("Data", func() {

	It("returns a the data type for the name", func() {
		d := NewData(strings.NewReader(""), data.AppUsageDataType)
		Expect(d.Name()).To(Equal(data.AppUsageDataType))
	})

	It("returns content for the data", func() {
		dataReader := strings.NewReader("best-data")
		d := NewData(dataReader, data.AppUsageDataType)
		Expect(d.Content()).To(Equal(dataReader))
	})

	It("returns json as data type", func() {
		d := NewData(nil, data.AppUsageDataType)
		Expect(d.MimeType()).To(Equal("application/json"))
	})

	It("returns the product type", func() {
		d := NewData(nil, data.AppUsageDataType)
		Expect(d.Type()).To(Equal(""))
	})

	It("returns the data type", func() {
		d := NewData(nil, data.AppUsageDataType)
		Expect(d.DataType()).To(Equal(data.AppUsageDataType))
	})

})
