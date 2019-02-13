package credhub

import (
	"io"
)

//go:generate counterfeiter . CredhubService
type CredhubService interface {
	Certificates() (io.Reader, error)
}

type DataCollector struct {
	credhubService CredhubService
}

func NewDataCollector(cs CredhubService) *DataCollector {
	return &DataCollector{
		credhubService: cs,
	}
}

func (dc *DataCollector) Collect() (Data, error) {
	certReader, err := dc.credhubService.Certificates()
	if err != nil {
		return Data{}, err
	}

	return NewData(certReader), nil
}
