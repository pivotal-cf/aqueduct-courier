package coreconsumption_test

import (
	"errors"
	"github.com/pivotal-cf/aqueduct-courier/coreconsumption"
	"github.com/pivotal-cf/aqueduct-courier/coreconsumption/coreconsumptionfakes"
	"github.com/pivotal-cf/om/api"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {

	var (
		service   *coreconsumption.Service
		requestor *coreconsumptionfakes.FakeRequestor
	)

	BeforeEach(func() {
		requestor = new(coreconsumptionfakes.FakeRequestor)
		service = &coreconsumption.Service{Requestor: requestor}
	})

	It("curl succeeds, returns valid CSV content", func() {
		// GIVEN
		body := &readerCloser{reader: strings.NewReader("1691950071,TAS,1,2\n1691953671,TAS,3,4\n")}
		requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

		// WHEN
		countReader, err := service.CoreCounts()

		// THEN
		Expect(err).NotTo(HaveOccurred())

		contentBytes, err := io.ReadAll(countReader)
		Expect(err).NotTo(HaveOccurred())

		contentString := string(contentBytes)
		expectedJson := `[
		  {
			"TimeReported": "2023-08-13T18:07:51Z",
			"PhysicalCoreCount": 1,
			"VirtualCoreCount": 2,
			"ProductIdentifier": "TAS"
		  },
		  {
			"TimeReported": "2023-08-13T19:07:51Z",
			"PhysicalCoreCount": 3,
			"VirtualCoreCount": 4,
			"ProductIdentifier": "TAS"
		  }
		]`
		Expect(contentString).Should(MatchJSON(expectedJson))
	})

	It("curl succeeds, but CSV is not valid", func() {
		// GIVEN
		body := &readerCloser{reader: strings.NewReader("not-a-timestamp,TAS,1,2")}
		requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, nil)

		// WHEN
		_, err := service.CoreCounts()

		// THEN
		Expect(err).To(HaveOccurred())
	})

	It("curl succeeds, but non-200 HTTP status is returned", func() {
		// GIVEN
		body := &readerCloser{reader: strings.NewReader("some-bad-result")}
		requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusConflict}, nil)

		// WHEN
		_, err := service.CoreCounts()

		// THEN
		Expect(err).To(HaveOccurred())
	})

	It("curl call files", func() {
		// GIVEN
		body := &readerCloser{reader: strings.NewReader("1691950071,TAS,1,2")}
		requestor.CurlReturns(api.RequestServiceCurlOutput{Body: body, StatusCode: http.StatusOK}, errors.New("some-error"))

		// WHEN
		_, err := service.CoreCounts()

		// THEN
		Expect(err).To(HaveOccurred())
	})

})

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
