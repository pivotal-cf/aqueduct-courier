package consumption_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/aqueduct-utils/data"

	"github.com/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/consumption"
	"github.com/pivotal-cf/aqueduct-courier/consumption/consumptionfakes"
)

var _ = Describe("DataCollector", func() {
	var (
		logger             *log.Logger
		bufferedOutput     *gbytes.Buffer
		consumptionService *consumptionfakes.FakeConsumptionService
		dataCollector      *DataCollector
	)

	BeforeEach(func() {
		bufferedOutput = gbytes.NewBuffer()
		logger = log.New(bufferedOutput, "", 0)
		consumptionService = new(consumptionfakes.FakeConsumptionService)
		dataCollector = NewDataCollector(*logger, consumptionService, "some-usage-url")
	})

	Describe("collect", func() {
		It("succeeds", func() {
			appUsagesReader := strings.NewReader("app instance data")
			serviceUsagesReader := strings.NewReader("service instance data")
			taskUsagesReader := strings.NewReader("task instance data")

			consumptionService.AppUsagesReturns(appUsagesReader, nil)
			consumptionService.ServiceUsagesReturns(serviceUsagesReader, nil)
			consumptionService.TaskUsagesReturns(taskUsagesReader, nil)

			collectedUsageData, err := dataCollector.Collect()
			Expect(err).ToNot(HaveOccurred())

			Expect(bufferedOutput).To(gbytes.Say("Collecting data from Usage Service at some-usage-url"))
			Expect(collectedUsageData).To(ConsistOf(
				NewData(appUsagesReader, data.AppUsageDataType),
				NewData(taskUsagesReader, data.TaskUsageDataType),
				NewData(serviceUsagesReader, data.ServiceUsageDataType)),
			)
		})

		It("returns an error when consumptionService.AppUsages errors", func() {
			consumptionService.AppUsagesReturns(nil, errors.New("Requesting things is hard"))
			collectedData, err := dataCollector.Collect()

			Expect(collectedData).To(BeEmpty())
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(AppUsageRequestError))))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when consumptionService.ServiceUsages errors", func() {
			consumptionService.ServiceUsagesReturns(nil, errors.New("Requesting things is hard"))
			collectedData, err := dataCollector.Collect()

			Expect(collectedData).To(BeEmpty())
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(ServiceUsageRequestError))))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when consumptionService.TaskUsages errors", func() {
			consumptionService.TaskUsagesReturns(nil, errors.New("Requesting things is hard"))
			collectedData, err := dataCollector.Collect()

			Expect(collectedData).To(BeEmpty())
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(TaskUsageRequestError))))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})
	})
})
