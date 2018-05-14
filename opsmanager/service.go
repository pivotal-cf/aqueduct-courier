package opsmanager

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pivotal-cf/om/api"
	"github.com/pkg/errors"
)

const (
	ProductResourcePathFormat          = "/api/v0/staged/products/%s/resources"
	DeployedProductsPath               = "/api/v0/deployed/products"
	VmTypesPath                        = "/api/v0/vm_types"
	DiagnosticReportPath               = "/api/v0/diagnostic_report"
	RequestFailureErrorFormat          = "Failed %s %s"
	RequestUnexpectedStatusErrorFormat = "%s %s returned with unexpected status %d"
)

type Service struct {
	Requestor Requestor
}

//go:generate counterfeiter . Requestor
type Requestor interface {
	Invoke(input api.RequestServiceInvokeInput) (api.RequestServiceInvokeOutput, error)
}

func (s *Service) DeployedProducts() (io.Reader, error) {
	return s.makeRequest(DeployedProductsPath)
}

func (s *Service) ProductResources(guid string) (io.Reader, error) {
	return s.makeRequest(fmt.Sprintf(ProductResourcePathFormat, guid))
}

func (s *Service) VmTypes() (io.Reader, error) {
	return s.makeRequest(VmTypesPath)
}

func (s *Service) DiagnosticReport() (io.Reader, error) {
	return s.makeRequest(DiagnosticReportPath)
}

func (s *Service) makeRequest(path string) (io.Reader, error) {
	input := api.RequestServiceInvokeInput{
		Path:   path,
		Method: http.MethodGet,
	}
	output, err := s.Requestor.Invoke(input)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, path))
	}
	if output.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf(RequestUnexpectedStatusErrorFormat, http.MethodGet, path, output.StatusCode))
	}
	return output.Body, nil
}
