package credhub

import (
	"io"
	"log"
)

//go:generate counterfeiter . CredhubService
type CredhubService interface {
	Certificates() (io.Reader, error)
}

type DataCollector struct {
	logger         log.Logger
	credhubService CredhubService
	credHubURL     string
}

func NewDataCollector(logger log.Logger, cs CredhubService, credHubURL string) *DataCollector {
	return &DataCollector{
		logger:         logger,
		credhubService: cs,
		credHubURL:     credHubURL,
	}
}

func (dc *DataCollector) Collect() (Data, error) {
	dc.logger.Printf("Collecting data from CredHub at %s", dc.credHubURL)
	certReader, err := dc.credhubService.Certificates()
	if err != nil {
		return Data{}, err
	}

	return NewData(certReader), nil
}
