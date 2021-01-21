package opsmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

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
	CertificatesPath            = "/api/v0/deployed/certificates"
	CertificateAuthoritiesPath  = "/api/v0/certificate_authorities"
	BoshCredentialsPath         = "/api/v0/deployed/director/credentials/bosh_commandline_credentials"

	ReadResponseBodyFailureFormat      = "Unable to read response from %s"
	InvalidResponseErrorFormat         = "Invalid response format for request to %s"
	RequestFailureErrorFormat          = "Failed %s %s"
	RequestUnexpectedStatusErrorFormat = "%s %s returned with unexpected status %d"
	UnmarshalResponseError             = "error unmarshalling response"
)

type Service struct {
	Requestor Requestor
}

type BoshCredential struct {
	ClientID     string
	ClientSecret string
	Host         string
}

type certificateAuthorities struct {
	CertificateAuthorities []certificateAuthority `json:"certificate_authorities"`
}

type certificateAuthority struct {
	GUID      string `json:"guid"`
	Issuer    string `json:"issuer"`
	CreatedOn string `json:"created_on"`
	ExpiresOn string `json:"expires_on"`
	Active    bool   `json:"active"`
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
	contents, err := s.makeRequest(InstallationsPath)
	if err != nil {
		return nil, err
	}

	var i installations
	if err := json.Unmarshal(contents, &i); err != nil {
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

func (s *Service) CertificateAuthorities() (io.Reader, error) {
	contents, err := s.makeRequest(CertificateAuthoritiesPath)
	if err != nil {
		return nil, err
	}

	var ca certificateAuthorities
	if err := json.Unmarshal(contents, &ca); err != nil {
		return nil, errors.Wrapf(err, InvalidResponseErrorFormat, CertificateAuthoritiesPath)
	}

	redactedContent, err := json.Marshal(ca)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(redactedContent), nil
}
func (s *Service) Certificates() (io.Reader, error) {
	return s.makeRequestReader(CertificatesPath)
}

func (s *Service) DeployedProducts() (io.Reader, error) {
	return s.makeRequestReader(DeployedProductsPath)
}

func (s *Service) ProductResources(guid string) (io.Reader, error) {
	return s.makeRequestReader(fmt.Sprintf(ProductResourcesPathFormat, guid))
}

func (s *Service) ProductProperties(guid string) (io.Reader, error) {
	productPropertiesPath := fmt.Sprintf(ProductPropertiesPathFormat, guid)
	contents, err := s.makeRequest(productPropertiesPath)
	if err != nil {
		return nil, err
	}

	var ps productProperties
	if err := json.Unmarshal(contents, &ps); err != nil {
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
	return s.makeRequestReader(VmTypesPath)
}

func (s *Service) DiagnosticReport() (io.Reader, error) {
	diagnosticReportBytes, err := s.makeRequest(DiagnosticReportPath)
	if err != nil {
		return nil, err
	}

	var diagnosticReportMap map[string]interface{}
	err = json.Unmarshal(diagnosticReportBytes, &diagnosticReportMap)
	if err != nil {
		return nil, errors.Wrap(err, UnmarshalResponseError)
	}

	val, _ := diagnosticReportMap["director_configuration"].(map[string]interface{})
	delete(val, "ntp_servers")

	redactedDiagnosticReport, err := json.Marshal(diagnosticReportMap)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(redactedDiagnosticReport), nil
}

func (s *Service) BoshCredentials() (BoshCredential, error) {
	credBytes, err := s.makeRequest(BoshCredentialsPath)
	if err != nil {
		return BoshCredential{}, err
	}

	var credentialMap map[string]string
	err = json.Unmarshal(credBytes, &credentialMap)
	if err != nil {
		return BoshCredential{}, errors.Errorf(InvalidResponseErrorFormat, BoshCredentialsPath)
	}

	credString := credentialMap["credential"]
	credStringParts := strings.Split(credString, " ")

	bCred := BoshCredential{}
	for _, part := range credStringParts {
		if strings.Contains(part, "=") {
			keyValueParts := strings.Split(part, "=")
			switch keyValueParts[0] {
			case "BOSH_CLIENT":
				bCred.ClientID = keyValueParts[1]
			case "BOSH_CLIENT_SECRET":
				bCred.ClientSecret = keyValueParts[1]
			case "BOSH_ENVIRONMENT":
				bCred.Host = keyValueParts[1]
			}
		}
	}

	return bCred, nil
}

func (s *Service) makeRequestReader(path string) (io.Reader, error) {
	content, err := s.makeRequest(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), nil
}

func (s *Service) makeRequest(path string) ([]byte, error) {
	input := api.RequestServiceCurlInput{
		Path:   path,
		Method: http.MethodGet,
		Headers: make(http.Header),
	}
	resp, err := s.Requestor.Curl(input)
	if err != nil {
		return nil, errors.Wrapf(err, RequestFailureErrorFormat, http.MethodGet, path)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf(RequestUnexpectedStatusErrorFormat, http.MethodGet, path, resp.StatusCode))
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, ReadResponseBodyFailureFormat, input.Path)
	}
	return contents, nil
}

func allowedPropertyType(propertyType string) bool {
	allowedTypes := []string{
		"integer",
		"boolean",
		"dropdown_select",
		"multi_select_options",
		"selector",
		"vm_type_dropdown",
		"disk_type_dropdown",
	}
	for _, p := range allowedTypes {
		if propertyType == p {
			return true
		}
	}
	return false
}
