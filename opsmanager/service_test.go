package opsmanager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"strings"

	"net/http"

	"fmt"

	. "github.com/pivotal-cf/aqueduct-collector/opsmanager"
	"github.com/pivotal-cf/aqueduct-collector/opsmanager/opsmanagerfakes"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("Service", func() {
	var (
		requestor *opsmanagerfakes.FakeRequester
		service   *Service
	)
	BeforeEach(func() {
		requestor = new(opsmanagerfakes.FakeRequester)
		service = &Service{
			Requestor: requestor,
		}
	})

	Describe("ProductResources", func() {
		var expectedProductPath string

		const productGUID = "product-guid"

		BeforeEach(func() {
			expectedProductPath = fmt.Sprintf(ProductResourcePathFormat, productGUID)
		})

		It("returns product resources content", func() {
			body := strings.NewReader("product-resources")

			requestor.InvokeReturns(api.RequestServiceInvokeOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.ProductResources(productGUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(body))

			Expect(requestor.InvokeCallCount()).To(Equal(1))
			input := requestor.InvokeArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceInvokeInput{Path: expectedProductPath, Method: http.MethodGet}))
		})

		It("returns an error when requestor errors", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.ProductResources(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, expectedProductPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.ProductResources(productGUID)
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, expectedProductPath, http.StatusBadGateway,
			)))
		})
	})
	Describe("DirectorProperties", func() {
		It("returns product resources content", func() {
			body := strings.NewReader("director-properties")

			requestor.InvokeReturns(api.RequestServiceInvokeOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.DirectorProperties()
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(body))
			Expect(requestor.InvokeCallCount()).To(Equal(1))
			input := requestor.InvokeArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceInvokeInput{Path: DirectorPropertiesPath, Method: http.MethodGet}))
		})

		It("returns an error when requestor errors", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.DirectorProperties()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, DirectorPropertiesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.DirectorProperties()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, DirectorPropertiesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("VmTypes", func() {
		It("returns product resources content", func() {
			body := strings.NewReader("vm-types")

			requestor.InvokeReturns(api.RequestServiceInvokeOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.VmTypes()
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(body))
			Expect(requestor.InvokeCallCount()).To(Equal(1))
			input := requestor.InvokeArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceInvokeInput{Path: VmTypesPath, Method: http.MethodGet}))
		})

		It("returns an error when requestor errors", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.VmTypes()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, VmTypesPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.VmTypes()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, VmTypesPath, http.StatusBadGateway,
			)))
		})
	})

	Describe("DiagnosticReport", func() {
		It("returns product resources content", func() {
			body := strings.NewReader("diagnostic-report")

			requestor.InvokeReturns(api.RequestServiceInvokeOutput{Body: body, StatusCode: http.StatusOK}, nil)

			actual, err := service.DiagnosticReport()
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(body))
			Expect(requestor.InvokeCallCount()).To(Equal(1))
			input := requestor.InvokeArgsForCall(0)
			Expect(input).To(Equal(api.RequestServiceInvokeInput{Path: DiagnosticReportPath, Method: http.MethodGet}))
		})

		It("returns an error when requestor errors", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusOK}, errors.New("Requesting things is hard"))

			actual, err := service.DiagnosticReport()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(
				fmt.Sprintf(RequestFailureErrorFormat, http.MethodGet, DiagnosticReportPath),
			)))
			Expect(err).To(MatchError(ContainSubstring("Requesting things is hard")))
		})

		It("returns an error when requestor returns a non 200 status code", func() {
			requestor.InvokeReturns(api.RequestServiceInvokeOutput{StatusCode: http.StatusBadGateway}, nil)

			actual, err := service.DiagnosticReport()
			Expect(actual).To(BeNil())
			Expect(err).To(MatchError(fmt.Sprintf(
				RequestUnexpectedStatusErrorFormat, http.MethodGet, DiagnosticReportPath, http.StatusBadGateway,
			)))
		})
	})
})
