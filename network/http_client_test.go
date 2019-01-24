package network_test

import (
	"net/http"
	"strings"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/network"
)

var _ = Describe("Client", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewTLSServer()
		server.RouteToHandler(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Do", func() {
		Context("when skipTLSVerification is set to false", func() {
			It("throws an error for invalid certificates", func() {
				client := NewClient(false)

				req, err := http.NewRequest(http.MethodGet, server.URL(), strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				_, err = client.Do(req)
				Expect(err.Error()).To(HaveSuffix("certificate signed by unknown authority"))
			})
		})

		Context("when skipTLSVerification is set to true", func() {
			It("does not verify certificates", func() {
				client := NewClient(true)

				req, err := http.NewRequest(http.MethodGet, server.URL(), strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				_, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
