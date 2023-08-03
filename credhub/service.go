package credhub

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	ListCertificatesError             = "Failed listing certificates from credhub"
	ListCertificatesReadError         = "Failed to read certificates response from credhub"
	ParseCertificatesError            = "Failed to parse certificates response from credhub"
	GetCertificateDataErrorFormat     = "Failed retrieving certificate %s from credhub"
	GetCertificateDataReadErrorFormat = "Failed to read certificate %s from credhub"
	CertificatePEMParseError          = "PEM decoding failed"
)

type Service struct {
	requestor credhubRequestor
}

//go:generate counterfeiter . credhubRequestor
type credhubRequestor interface {
	Request(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error)
}

func NewCredhubService(requestor credhubRequestor) *Service {
	return &Service{requestor: requestor}
}

func (s *Service) Certificates() (io.Reader, error) {
	query := url.Values{}
	resp, err := s.requestor.Request(http.MethodGet, "/api/v1/certificates", query, nil, true)
	if err != nil {
		return nil, errors.Wrap(err, ListCertificatesError)
	}

	defer resp.Body.Close()

	certificatesContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, ListCertificatesReadError)
	}

	var parsedCertificates map[string][]struct{ Name string }
	err = json.Unmarshal(certificatesContent, &parsedCertificates)
	if err != nil {
		return nil, errors.Wrap(err, ParseCertificatesError)
	}

	var certificateInfos []map[string]string
	for _, certMap := range parsedCertificates["certificates"] {
		query := url.Values{}
		query.Set("name", certMap.Name)

		resp, err := s.requestor.Request(http.MethodGet, "/api/v1/data", query, nil, true)
		if err != nil {
			return nil, errors.Wrapf(err, GetCertificateDataErrorFormat, certMap.Name)
		}

		cert, err := parseCertFromDataResponse(resp.Body)
		if err != nil {
			return nil, errors.Wrapf(err, GetCertificateDataReadErrorFormat, certMap.Name)
		}
		resp.Body.Close()

		certificateInfos = append(certificateInfos, map[string]string{
			"name":       certMap.Name,
			"not_before": cert.NotBefore.Format(time.RFC3339),
			"not_after":  cert.NotAfter.Format(time.RFC3339),
			"issuer":     cert.Issuer.String(),
		})
	}

	jsonBytes, err := json.Marshal(map[string][]map[string]string{"credhub_certificates": certificateInfos})
	return bytes.NewReader(jsonBytes), nil
}

func parseCertFromDataResponse(body io.Reader) (*x509.Certificate, error) {
	dataContent, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var parsedDataResponse map[string][]struct {
		Value struct {
			Certificate string
		}
	}

	err = json.Unmarshal(dataContent, &parsedDataResponse)
	if err != nil {
		return nil, err
	}

	certPEMString := parsedDataResponse["data"][0].Value.Certificate

	block, _ := pem.Decode([]byte(certPEMString))
	if block == nil {
		return nil, errors.New(CertificatePEMParseError)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
