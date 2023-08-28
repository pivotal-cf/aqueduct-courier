package coreconsumption_test

import (
	"errors"
	"log"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-cf/aqueduct-courier/coreconsumption"
	"github.com/pivotal-cf/aqueduct-courier/coreconsumption/coreconsumptionfakes"
)

var _ = Describe("DataCollector", func() {

	var (
		logger         *log.Logger
		bufferedOutput *gbytes.Buffer
		omURL          string
		collector      *DataCollector
		omService      *coreconsumptionfakes.FakeOmService
	)

	BeforeEach(func() {
		bufferedOutput = gbytes.NewBuffer()
		logger = log.New(bufferedOutput, "", 0)
		omURL = "some-opsmanager-url"
		omService = new(coreconsumptionfakes.FakeOmService)
		collector = NewDataCollector(logger, omService, omURL)
	})

	It("retriever succeeds", func() {
		// GIVEN
		omService.CoreCountsReturns(strings.NewReader("some-csv-content"), nil)

		// WHEN
		data, err := collector.Collect()

		// THEN
		Expect(err).NotTo(HaveOccurred())
		Expect(data).Should(HaveLen(1))
	})

	It("retriever fails", func() {
		// GIVEN
		omService.CoreCountsReturns(nil, errors.New("some-error"))

		// WHEN
		data, err := collector.Collect()

		// THEN
		Expect(err).To(HaveOccurred())
		Expect(data).Should(HaveLen(0))
	})
})
