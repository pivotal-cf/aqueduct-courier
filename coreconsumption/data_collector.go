package coreconsumption

import (
	"fmt"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"
	"github.com/pkg/errors"
	"io"
	"log"
)

const (
	RequestorFailureErrorFormat = "Failed retrieving %s %s"
)

//go:generate counterfeiter . OmService
type OmService interface {
	CoreCounts() (io.Reader, error)
}

type dataRetriever func() (io.Reader, error)

type DataCollector struct {
	logger        log.Logger
	omService     OmService
	opsManagerURL string
}

func NewDataCollector(logger log.Logger, oms OmService, omURL string) *DataCollector {
	return &DataCollector{
		logger:        logger,
		omService:     oms,
		opsManagerURL: omURL,
	}
}

func (dc *DataCollector) Collect() ([]Data, error) {
	dc.logger.Printf("Collecting data from Operations Manager at %s", dc.opsManagerURL)

	d, err := appendRetrievedData(dc.omService.CoreCounts, "", collector_tar.CoreCountsDataType)
	if err != nil {
		return []Data{}, err
	}

	return d, nil
}

func appendRetrievedData(retriever dataRetriever, productType, dataType string) ([]Data, error) {
	var d []Data
	output, err := retriever()
	if err != nil {
		return d, errors.Wrap(err, fmt.Sprintf(RequestorFailureErrorFormat, productType, dataType))
	}

	return append(d, NewData(output, productType, dataType)), nil
}
