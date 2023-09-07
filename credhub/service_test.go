package credhub_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/credhub"
	"github.com/pivotal-cf/aqueduct-courier/credhub/credhubfakes"
	"github.com/pkg/errors"
)

var _ = Describe("Service", func() {

	It("returns the parsed certificate information from credhub", func() {
		expectedNotBefore1 := time.Now().UTC()
		expectedNotAfter1 := expectedNotBefore1.Add(1234 * time.Second)
		cert1 := makeCert(expectedNotBefore1, expectedNotAfter1, "org1-name")

		expectedNotBefore2 := time.Now().Add(99 * time.Second).UTC()
		expectedNotAfter2 := expectedNotBefore2.Add(1234 * time.Second)
		cert2 := makeCert(expectedNotBefore2, expectedNotAfter2, "org2-name")

		certListResponse := makeCertListResponse("cert1-name-path", "cert2-name-path")

		cert1DataResponseStruct := map[string][]map[string]map[string]string{
			"data": {{"value": {"certificate": cert1}}},
		}
		cert1DataResponse, err := json.Marshal(cert1DataResponseStruct)
		Expect(err).NotTo(HaveOccurred())

		cert2DataResponseStruct := map[string][]map[string]map[string]string{
			"data": {{"value": {"certificate": cert2}}},
		}
		cert2DataResponse, err := json.Marshal(cert2DataResponseStruct)
		Expect(err).NotTo(HaveOccurred())

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		certificatesBody := &readerCloser{reader: bytes.NewReader(certListResponse)}
		cert1Body := &readerCloser{reader: bytes.NewReader(cert1DataResponse)}
		cert2Body := &readerCloser{reader: bytes.NewReader(cert2DataResponse)}

		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			Expect(method).To(Equal(http.MethodGet))
			Expect(checkServerErr).To(BeTrue())
			Expect(body).To(BeNil())
			switch pathStr {
			case "/api/v1/certificates":
				Expect(query).To(Equal(url.Values{}))
				response.Body = certificatesBody
			case "/api/v1/data":
				certPath := query.Get("name")
				switch certPath {
				case "cert1-name-path":
					response.Body = cert1Body
				case "cert2-name-path":
					response.Body = cert2Body
				default:
					Fail(fmt.Sprintf("Unexpected cert path %s", certPath))
				}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}

		service := NewCredhubService(credhubRequestor)

		reader, err := service.Certificates()
		Expect(err).NotTo(HaveOccurred())

		certContent, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())

		Expect(certificatesBody.isClosed).To(BeTrue())
		Expect(cert1Body.isClosed).To(BeTrue())
		Expect(cert2Body.isClosed).To(BeTrue())

		var credhubCertificates map[string][]map[string]string
		Expect(json.Unmarshal(certContent, &credhubCertificates)).To(Succeed())
		Expect(credhubCertificates["credhub_certificates"]).To(Equal([]map[string]string{
			{
				"name":       "cert1-name-path",
				"not_before": expectedNotBefore1.Format(time.RFC3339),
				"not_after":  expectedNotAfter1.Format(time.RFC3339),
				"issuer":     "O=org1-name,C=Melchizedek",
			},
			{
				"name":       "cert2-name-path",
				"not_before": expectedNotBefore2.Format(time.RFC3339),
				"not_after":  expectedNotAfter2.Format(time.RFC3339),
				"issuer":     "O=org2-name,C=Melchizedek",
			},
		}))
	})

	It("returns an error when fetching a list of certificates fails", func() {
		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			switch pathStr {
			case "/api/v1/certificates":
				return nil, errors.New("requesting stuff is hard")
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return nil, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("requesting stuff is hard")))
		Expect(err).To(MatchError(ContainSubstring(ListCertificatesError)))
	})

	It("returns an error if reading the certificate listing fails", func() {
		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: &badReader{}}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("Reading is hard")))
		Expect(err).To(MatchError(ContainSubstring(ListCertificatesReadError)))
	})

	It("returns an error if unmarshalling the certificate list fails", func() {
		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: strings.NewReader("Simply not JSON!")}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring(ParseCertificatesError)))
	})

	It("returns an error if fetching certificate data fails", func() {
		certListResponse := makeCertListResponse("cert1-name-path")

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: bytes.NewReader(certListResponse)}
			case "/api/v1/data":
				return nil, errors.New("requesting data stuff is hard")
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("requesting data stuff is hard")))
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(GetCertificateDataErrorFormat, "cert1-name-path"))))
	})

	It("returns an error if reading the certificate data fails", func() {
		certListResponse := makeCertListResponse("cert1-name-path")

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: bytes.NewReader(certListResponse)}
			case "/api/v1/data":
				response.Body = &readerCloser{reader: &badReader{}}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("Reading is hard")))
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(GetCertificateDataReadErrorFormat, "cert1-name-path"))))
	})

	It("returns an error if unmarshalling the certificate data fails", func() {
		certListResponse := makeCertListResponse("cert1-name-path")

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: bytes.NewReader(certListResponse)}
			case "/api/v1/data":
				response.Body = &readerCloser{reader: strings.NewReader("totally not json")}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(GetCertificateDataReadErrorFormat, "cert1-name-path"))))
	})

	It("returns an error if pem decoding the certificate data fails", func() {
		certListResponse := makeCertListResponse("cert1-name-path")

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: bytes.NewReader(certListResponse)}
			case "/api/v1/data":
				response.Body = &readerCloser{reader: strings.NewReader(`{"data": [{}]}`)}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err := service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring(CertificatePEMParseError)))
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(GetCertificateDataReadErrorFormat, "cert1-name-path"))))
	})

	It("returns an error if parsing the certificate data fails", func() {
		certListResponse := makeCertListResponse("cert1-name-path")

		buffer := bytes.NewBuffer([]byte{})
		Expect(pem.Encode(buffer, &pem.Block{Type: "CERTIFICATE", Bytes: []byte("invalid-cert-content")})).To(Succeed())

		parsedDataResponseStruct := map[string][]map[string]map[string]string{
			"data": {{"value": {"certificate": buffer.String()}}},
		}
		parsedDataResponse, err := json.Marshal(parsedDataResponseStruct)
		Expect(err).NotTo(HaveOccurred())

		credhubRequestor := new(credhubfakes.FakeCredhubRequestor)
		credhubRequestor.RequestStub = func(method string, pathStr string, query url.Values, body interface{}, checkServerErr bool) (*http.Response, error) {
			response := http.Response{}
			switch pathStr {
			case "/api/v1/certificates":
				response.Body = &readerCloser{reader: bytes.NewReader(certListResponse)}
			case "/api/v1/data":
				response.Body = &readerCloser{reader: bytes.NewReader(parsedDataResponse)}
			default:
				Fail(fmt.Sprintf("Unexpected request path %s", pathStr))
			}
			return &response, nil
		}
		service := NewCredhubService(credhubRequestor)

		_, err = service.Certificates()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(GetCertificateDataReadErrorFormat, "cert1-name-path"))))
	})
})

func makeCertListResponse(certNames ...string) []byte {
	certListResponseStruct := map[string][]map[string]interface{}{}
	certListResponseStruct["certificates"] = []map[string]interface{}{}
	for _, certName := range certNames {
		certListResponseStruct["certificates"] = append(certListResponseStruct["certificates"], map[string]interface{}{
			"name":      certName,
			"unwanted":  true,
			"no-matter": []string{"ever"},
		})
	}
	certListResponse, err := json.Marshal(certListResponseStruct)
	Expect(err).NotTo(HaveOccurred())
	return certListResponse
}

type badReader struct{}

func (r *badReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("Reading is hard")
}

func makeCert(notBefore, notAfter time.Time, issuingOrg string) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())

	issuerCert := x509.Certificate{
		SerialNumber: big.NewInt(42),
		Subject: pkix.Name{
			Country:      []string{"Melchizedek"},
			Organization: []string{issuingOrg},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		Issuer: pkix.Name{
			Country:      []string{"Melchizedek"},
			Organization: []string{issuingOrg},
		},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &issuerCert, &privateKey.PublicKey, privateKey)
	Expect(err).NotTo(HaveOccurred())

	certOut := bytes.NewBuffer([]byte{})
	Expect(pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})).To(Succeed())

	return certOut.String()
}

type readerCloser struct {
	reader   io.Reader
	isClosed bool
}

func (rc *readerCloser) Read(p []byte) (n int, err error) {
	if rc.isClosed {
		return 0, errors.New("Read called on closed body")
	}
	return rc.reader.Read(p)
}

func (rc *readerCloser) Close() error {
	rc.isClosed = true
	return nil
}
