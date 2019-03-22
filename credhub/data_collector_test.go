package credhub_test

import (
	"log"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/aqueduct-courier/credhub/credhubfakes"
	"github.com/pkg/errors"
)

var _ = Describe("DataCollector", func() {

	var (
		logger         *log.Logger
		bufferedOutput *gbytes.Buffer
		credHubURL     string
	)

	BeforeEach(func() {
		bufferedOutput = gbytes.NewBuffer()
		logger = log.New(bufferedOutput, "", 0)
		credHubURL = "some-credhub-url"
	})

	It("returns data using the credhub service", func() {
		certificatesReader := strings.NewReader("certificates data reader")
		credHubService := new(credhubfakes.FakeCredhubService)
		credHubService.CertificatesReturns(certificatesReader, nil)
		collector := NewDataCollector(*logger, credHubService, credHubURL)

		data, err := collector.Collect()
		Expect(err).NotTo(HaveOccurred())
		Expect(bufferedOutput).To(gbytes.Say("Collecting data from CredHub at some-credhub-url"))
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
