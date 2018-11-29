package credhub_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/aqueduct-courier/credhub/credhubfakes"
	"github.com/pkg/errors"
	"strings"
)

var _ = Describe("DataCollector", func() {

	It("returns data using the credhub service", func() {
		certificatesReader := strings.NewReader("certificates data reader")
		credhubService := new(credhubfakes.FakeCredhubService)
		credhubService.CertificatesReturns(certificatesReader, nil)
		collector := NewDataCollector(credhubService)

		data, err := collector.Collect()
		Expect(err).NotTo(HaveOccurred())

		Expect(data).To(Equal(NewData(certificatesReader)))
	})

	It("returns an error when collecting certificates fails", func() {
		credhubService := new(credhubfakes.FakeCredhubService)
		credhubService.CertificatesReturns(nil, errors.New("collecting certificates is hard"))
		collector := NewDataCollector(credhubService)

		_, err := collector.Collect()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("collecting certificates is hard"))
	})

})
