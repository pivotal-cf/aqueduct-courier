package opsmanager_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager/opsmanagerfakes"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("Service", func() {
	var (
		requestor *opsmanagerfakes.FakeRequestor
		service   *Service
	)
	BeforeEach(func() {
		requestor = new(opsmanagerfakes.FakeRequestor)
		service = &Service{
			Requestor: requestor,
		}
	})

	Describe("DeployedProducts", func() {
		It("returns deployed products content", func() {
			body := &readerCloser{reader: strings.NewReader("deployed-products")}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.DeployedProducts()
			Expect(err).NotTo(HaveOccurred())
			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte("deployed-products")))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    DeployedProductsPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.DeployedProducts()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, DeployedProductsPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.DeployedProducts()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, DeployedProductsPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("ProductResources", func() {
		var expectedProductPath string

		const productGUID = "product-guid"

		BeforeEach(func() {
			expectedProductPath = fmt.Sprintf(ProductResourcesPathFormat, productGUID)
		})

		It("returns product resources content", func() {
			body := &readerCloser{reader: strings.NewReader("product-resources")}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.ProductResources(productGUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte("product-resources")))

			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    expectedProductPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.ProductResources(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, expectedProductPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.ProductResources(productGUID)
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, expectedProductPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("ProductProperties", func() {
		var expectedProductPropertiesPath string

		const productGUID = "product-guid"

		BeforeEach(func() {
			expectedProductPropertiesPath = fmt.Sprintf(ProductPropertiesPathFormat, productGUID)
		})

		It("returns product properties with specific safe types", func() {
			properties := map[string]map[string]map[string]interface{}{
				"properties": {
					"path.to1": {
						"type":         "boolean",
						"value":        true,
						"otherKey":     1234,
						"configurable": true,
						"credential":   true,
						"optional":     true,
					},
					"path.to2": {
						"type":         "integer",
						"value":        "2",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"path.to3": {
						"type":         "dropdown_select",
						"value":        "whatever",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"path.to4": {
						"type":         "multi_select_options",
						"value":        "true",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"path.to5": {
						"type":         "selector",
						"value":        "selected_option_words",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"path.to6": {
						"type":         "vm_type_dropdown",
						"value":        "selected_vm_type",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"path.to7": {
						"type":         "disk_type_dropdown",
						"value":        "selected_disk_type",
						"configurable": false,
						"credential":   false,
						"optional":     false,
					},
					"remove1": {
						"type":  "unknown",
						"value": "other stuff",
					},
					"remove2": {
						"type":  "collection",
						"value": "stuff",
					},
				},
			}
			propertiesJson, err := json.Marshal(properties)
			Expect(err).NotTo(HaveOccurred())
			body := ioutil.NopCloser(bytes.NewReader(propertiesJson))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			expectedProperties := properties
			delete(expectedProperties["properties"]["path.to1"], "otherKey")
			delete(expectedProperties["properties"], "remove1")
			delete(expectedProperties["properties"], "remove2")

			actual, err := service.ProductProperties(productGUID)
			Expect(err).NotTo(HaveOccurred())
			actualContent, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())

			var actualProperties map[string]map[string]map[string]interface{}
			Expect(json.Unmarshal(actualContent, &actualProperties)).To(Succeed())
			Expect(actualProperties).To(Equal(expectedProperties))

			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    expectedProductPropertiesPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("errors if the contents cannot be read from the response", func() {
			badReader := new(opsmanagerfakes.FakeReader)
			badReader.ReadReturns(0, errors.New("Reading things is hard"))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: ioutil.NopCloser(badReader), StatusCode: http.StatusOK}, nil)

			actual, err := service.ProductProperties(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(ReadResponseBodyFailureFormat, expectedProductPropertiesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Reading things is hard")))
		})

		It("errors if the contents are not json", func() {
			body := ioutil.NopCloser(strings.NewReader(`you-thought-this-was-json`))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.ProductProperties(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(InvalidResponseErrorFormat, expectedProductPropertiesPath),
			)))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.ProductProperties(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, expectedProductPropertiesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.ProductProperties(productGUID)
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, expectedProductPropertiesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("VmTypes", func() {
		It("returns product resources content", func() {
			body := &readerCloser{reader: strings.NewReader("vm-types")}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.VmTypes()
			Expect(err).NotTo(HaveOccurred())
			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte("vm-types")))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    VmTypesPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.VmTypes()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, VmTypesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.VmTypes()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, VmTypesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("DiagnosticReport", func() {
		It("returns product resources content without ntp server", func() {

			rawDiagnosticReportContents := `{
"other-valid-key": true,
"director_configuration": {
    "bosh_recreate_on_next_deploy": false,
    "resurrector_enabled": false,
    "blobstore_type": "local",
    "max_threads": null,
    "database_type": "internal",
	"ntp_servers": [
      "169.254.169.254"
    ]}}`

			expectedDiagnosticReport := map[string]interface{}{
				"other-valid-key": true,
				"director_configuration": map[string]interface{}{
					"bosh_recreate_on_next_deploy": false,
					"resurrector_enabled":          false,
					"blobstore_type":               "local",
					"max_threads":                  nil,
					"database_type":                "internal"},
			}

			body := &readerCloser{reader: strings.NewReader(rawDiagnosticReportContents)}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.DiagnosticReport()
			Expect(err).NotTo(HaveOccurred())
			Expect(body.isClosed).To(BeTrue())
			actualContent, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())

			var actualDiagnosticReport map[string]interface{}
			Expect(json.Unmarshal(actualContent, &actualDiagnosticReport)).To(Succeed())
			Expect(actualDiagnosticReport).To(Equal(expectedDiagnosticReport))

			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    DiagnosticReportPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when unmarshalling the response errors", func() {
			body := &readerCloser{reader: bytes.NewReader([]byte(`something-invalid`))}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK, Body: body}, nil)

			actual, err := service.DiagnosticReport()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(UnmarshalResponseError)))
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.DiagnosticReport()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, DiagnosticReportPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.DiagnosticReport()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, DiagnosticReportPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("Installations", func() {
		It("removes user names from the installation content and returns the rest", func() {
			body := &readerCloser{reader: strings.NewReader(`{"installations": [{"user_name": "foo", "other": 42}, {"user_name": "bar", "other": 24}]}`)}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.Installations()
			Expect(err).NotTo(HaveOccurred())
			actualContent, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(actualContent)).To(Equal(`{"installations":[{"other":42},{"other":24}]}`))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    InstallationsPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("errors if the contents cannot be read from the response", func() {
			badReader := new(opsmanagerfakes.FakeReader)
			badReader.ReadReturns(0, errors.New("Reading things is hard"))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: ioutil.NopCloser(badReader), StatusCode: http.StatusOK}, nil)

			actual, err := service.Installations()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(ReadResponseBodyFailureFormat, InstallationsPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Reading things is hard")))
		})

		It("errors if the contents are not json", func() {
			body := &readerCloser{reader: strings.NewReader(`you-thought-this-was-json`)}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.Installations()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(InvalidResponseErrorFormat, InstallationsPath),
			)))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.Installations()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, InstallationsPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.Installations()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, InstallationsPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("Certificates", func() {
		It("returns deployed certificates content", func() {
			body := &readerCloser{reader: strings.NewReader(`{"certificates":[{"keys": "for-certs"}]}`)}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.Certificates()
			Expect(err).NotTo(HaveOccurred())
			actualContent, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actualContent)).To(Equal(`{"certificates":[{"keys": "for-certs"}]}`))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    CertificatesPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.Certificates()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, CertificatesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.Certificates()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, CertificatesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("CertificateAuthorities", func() {
		It("returns deployed certificates content, removing unknown keys", func() {
			body := &readerCloser{reader: strings.NewReader(`{
"certificate_authorities":[{
	"guid": "f7bc18f34f2a7a9403c3",
	"issuer": "VMware",
	"created_on": "2017-02-09",
	"expires_on": "2021-01-10",
	"active": true,
	"cert_pem": "should not be here",
	"nats_cert_pem": "should not be here",
	"random_key": "should not be here"
}]}`)}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.CertificateAuthorities()
			Expect(err).NotTo(HaveOccurred())
			actualContent, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			var actualCertAuths map[string][]interface{}
			err = json.Unmarshal(actualContent, &actualCertAuths)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualCertAuths["certificate_authorities"]).To(ConsistOf(map[string]interface{}{
				"guid":       "f7bc18f34f2a7a9403c3",
				"issuer":     "VMware",
				"created_on": "2017-02-09",
				"expires_on": "2021-01-10",
				"active":     true,
			}))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    CertificateAuthoritiesPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("errors if the contents cannot be read from the response", func() {
			badReader := new(opsmanagerfakes.FakeReader)
			badReader.ReadReturns(0, errors.New("Reading things is hard"))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: ioutil.NopCloser(badReader), StatusCode: http.StatusOK}, nil)

			actual, err := service.CertificateAuthorities()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(ReadResponseBodyFailureFormat, CertificateAuthoritiesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Reading things is hard")))
		})

		It("errors if the contents are not json", func() {
			body := &readerCloser{reader: strings.NewReader(`you-thought-this-was-json`)}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.CertificateAuthorities()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(InvalidResponseErrorFormat, CertificateAuthoritiesPath),
			)))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.CertificateAuthorities()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, CertificateAuthoritiesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.CertificateAuthorities()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, CertificateAuthoritiesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("BoshCredentials", func() {
		It("returns the bosh credentials content", func() {
			body := &readerCloser{reader: strings.NewReader(`{ "credential": "BOSH_CLIENT=best_client BOSH_CLIENT_SECRET=best_secret BOSH_CA_CERT=/cool/path BOSH_ENVIRONMENT=10.9.8.7 bosh "}`)}
			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.BoshCredentials()
			Expect(err).NotTo(HaveOccurred())

			Expect(actual.ClientID).To(Equal("best_client"))
			Expect(actual.ClientSecret).To(Equal("best_secret"))
			Expect(actual.Host).To(Equal("10.9.8.7"))

			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    BoshCredentialsPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("errors if the contents cannot be read from the response", func() {
			badReader := new(opsmanagerfakes.FakeReader)
			badReader.ReadReturns(0, errors.New("Reading things is hard"))

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: ioutil.NopCloser(badReader), StatusCode: http.StatusOK}, nil)

			actual, err := service.BoshCredentials()
			Expect(actual).To(Equal(BoshCredential{}))
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(ReadResponseBodyFailureFormat, BoshCredentialsPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Reading things is hard")))
		})

		It("errors if the contents are not json", func() {
			body := &readerCloser{reader: strings.NewReader(`you-thought-this-was-json`)}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.BoshCredentials()
			Expect(actual).To(Equal(BoshCredential{}))
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(InvalidResponseErrorFormat, BoshCredentialsPath),
			)))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.BoshCredentials()
			Expect(actual).To(Equal(BoshCredential{}))
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, BoshCredentialsPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.BoshCredentials()
			Expect(actual).To(Equal(BoshCredential{}))
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, BoshCredentialsPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("PendingChanges", func() {
		It("returns product resources content", func() {
			body := &readerCloser{reader: strings.NewReader("pending_changes")}

			requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.PendingChanges()
			Expect(err).NotTo(HaveOccurred())
			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte("pending_changes")))
			Expect(requestor.CurlCallCount()).To(Equal(1))
			input := requestor.CurlArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceCurlInput{
				Path:    PendingChangesPath,
				Method:  http.MethodGet,
				Headers: make(http.Header),
			}))
		})

		It("returns an error when requestor errors", func() {
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusOK}, errors.New("I had trouble detecting stuff"))

			actual, err := service.PendingChanges()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, PendingChangesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("I had trouble detecting stuff")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			body := &readerCloser{}
			requestor.CurlReturns(api.RequestServiceCurlOutput{StatusCode: http.StatusBadGateway, Body: body}, nil)

			actual, err := service.PendingChanges()
			Expect(actual).To(BeNil())
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, PendingChangesPath, http.StatusBadGateway,
			)))
		})
	})
})

//go:generate counterfeiter . reader
type reader interface {
	io.Reader
}

type readerCloser struct {
	reader   io.Reader
	isClosed bool
}

func (rc *readerCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *readerCloser) Close() error {
	rc.isClosed = true
	return nil
}
