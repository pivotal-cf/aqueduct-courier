package consumption

import (
	"io"
	"log"

	"github.com/pivotal-cf/telemetry-utils/collector_tar"
	"github.com/pkg/errors"
)

const (
	AppUsageRequestError     = "Failed retrieving app usage data"
	ServiceUsageRequestError = "Failed retrieving service usage data"
	TaskUsageRequestError    = "Failed retrieving task usage data"
	SystemReportPathPrefix   = "system_report"
)

//go:generate counterfeiter . consumptionService
type consumptionService interface {
	AppUsages() (io.Reader, error)
	ServiceUsages() (io.Reader, error)
	TaskUsages() (io.Reader, error)
}

type DataCollector struct {
	logger             *log.Logger
	consumptionService consumptionService
	usageServiceURL    string
}

func NewDataCollector(logger *log.Logger, cs consumptionService, usageServiceURL string) *DataCollector {
	return &DataCollector{
		logger:             logger,
		consumptionService: cs,
		usageServiceURL:    usageServiceURL,
	}
}

func (dc *DataCollector) Collect() ([]Data, error) {
	dc.logger.Printf("Collecting data from Usage Service at %s", dc.usageServiceURL)

	appUsagesDataReader, err := dc.consumptionService.AppUsages()
	if err != nil {
		return []Data{}, errors.Wrap(err, AppUsageRequestError)
	}

	serviceUsagesDataReader, err := dc.consumptionService.ServiceUsages()
	if err != nil {
		return []Data{}, errors.Wrap(err, ServiceUsageRequestError)
	}

	taskUsagesDataReader, err := dc.consumptionService.TaskUsages()
	if err != nil {
		return []Data{}, errors.Wrap(err, TaskUsageRequestError)
	}

	return []Data{
		NewData(appUsagesDataReader, collector_tar.AppUsageDataType),
		NewData(serviceUsagesDataReader, collector_tar.ServiceUsageDataType),
		NewData(taskUsagesDataReader, collector_tar.TaskUsageDataType),
	}, nil
}
