package opsmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pivotal-cf/om/api"
	"github.com/pkg/errors"
)

const (
	ProductResourcesPathFormat  = "/api/v0/staged/products/%s/resources"
	ProductPropertiesPathFormat = "/api/v0/staged/products/%s/properties"
	InstallationsPath           = "/api/v0/installations"
	DeployedProductsPath        = "/api/v0/deployed/products"
	VmTypesPath                 = "/api/v0/vm_types"
	DiagnosticReportPath        = "/api/v0/diagnostic_report"

	ReadResponseBodyFailureFormat      = "Unable to read response from %s"
	InvalidResponseErrorFormat         = "Invalid response format for request to %s"
	RequestFailureErrorFormat          = "Failed %s %s"
	RequestUnexpectedStatusErrorFormat = "%s %s returned with unexpected status %d"
)

type Service struct {
	Requestor Requestor
}

type installations struct {
	Installations []map[string]interface{} `json:"installations"`
}

type productProperties struct {
	Properties map[string]property `json:"properties"`
}

type property struct {
	Type         string      `json:"type"`
	Value        interface{} `json:"value"`
	Configurable bool        `json:"configurable"`
	Credential   bool        `json:"credential"`
	Optional     bool        `json:"optional"`
}

//go:generate counterfeiter . Requestor
type Requestor interface {
	Curl(input api.RequestServiceCurlInput) (api.RequestServiceCurlOutput, error)
}

func (s *Service) Installations() (io.Reader, error) {
	contentReader, err := s.makeRequest(InstallationsPath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return nil, errors.Wrapf(err, ReadResponseBodyFailureFormat, InstallationsPath)
	}

	var i installations
	if err := json.Unmarshal([]byte(contents), &i); err != nil {
		return nil, errors.Wrapf(err, InvalidResponseErrorFormat, InstallationsPath)
	}
	for _, installation := range i.Installations {
		delete(installation, "user_name")
	}

	redactedContent, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(redactedContent), nil
}

func (s *Service) DeployedProducts() (io.Reader, error) {
	return s.makeRequest(DeployedProductsPath)
}

func (s *Service) ProductResources(guid string) (io.Reader, error) {
	return s.makeRequest(fmt.Sprintf(ProductResourcesPathFormat, guid))
}

func (s *Service) ProductProperties(guid string) (io.Reader, error) {
	productPropertiesPath := fmt.Sprintf(ProductPropertiesPathFormat, guid)
	contentReader, err := s.makeRequest(productPropertiesPath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(contentReader)
	if err != nil {
		return nil, errors.Wrapf(err, ReadResponseBodyFailureFormat, productPropertiesPath)
	}

	var ps productProperties
	if err := json.Unmarshal([]byte(contents), &ps); err != nil {
		return nil, errors.Wrapf(err, InvalidResponseErrorFormat, productPropertiesPath)
	}
	for propertyName, property := range ps.Properties {
		if !allowedPropertyType(property.Type) {
			delete(ps.Properties, propertyName)
		}
	}

	redactedContent, err := json.Marshal(ps)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(redactedContent), nil
}

func (s *Service) VmTypes() (io.Reader, error) {
	return s.makeRequest(VmTypesPath)
}

func (s *Service) DiagnosticReport() (io.Reader, error) {
	return s.makeRequest(DiagnosticReportPath)
}

func (s *Service) makeRequest(path string) (io.Reader, error) {
	input := api.RequestServiceCurlInput{
		Path:   path,
		Method: http.MethodGet,
	}
	output, err := s.Requestor.Curl(input)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, path))
	}
	if output.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf(RequestUnexpectedStatusErrorFormat, http.MethodGet, path, output.StatusCode))
	}
	return output.Body, nil
}

func allowedPropertyType(propertyType string) bool {
	for _, p := range []string{"integer", "boolean", "dropdown_select", "multi_select_options", "selector"} {
		if propertyType == p {
			return true
		}
	}
	return false
}
