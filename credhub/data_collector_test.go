package credhub_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/aqueduct-courier/credhub/credhubfakes"
	"github.com/pkg/errors"
	"log"
	"strings"
)

var _ = Describe("DataCollector", func() {

	var (
		logger *log.Logger
		credHubURL string
	)

	BeforeEach(func(){
		logger = log.New(GinkgoWriter, "", 0)
	})

	It("returns data using the credhub service", func() {
		certificatesReader := strings.NewReader("certificates data reader")
		credHubService := new(credhubfakes.FakeCredhubService)
		credHubService.CertificatesReturns(certificatesReader, nil)
		collector := NewDataCollector(*logger, credHubService, credHubURL)

		data, err := collector.Collect()
		Expect(err).NotTo(HaveOccurred())

		Expect(data).To(Equal(NewData(certificatesReader)))
	})

	It("returns an error when collecting certificates fails", func() {
		credHubService := new(credhubfakes.FakeCredhubService)
		credHubService.CertificatesReturns(nil, errors.New("collecting certificates is hard"))
		collector := NewDataCollector(*logger, credHubService, credHubURL)

		_, err := collector.Collect()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("collecting certificates is hard"))
	})

})
