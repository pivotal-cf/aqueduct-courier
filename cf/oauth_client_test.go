package cf_test

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/cf"
)

var _ = Describe("OAuthClient", func() {
	var (
		receivedRequest []byte
		authHeader      string
		oauthServer     *httptest.Server
		server          *httptest.Server
		accessURL       string
	)

	BeforeEach(func() {
		oauthServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/oauth/token" {
				var err error
				receivedRequest, err = httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())

				w.Header().Set("Content-Type", "application/json")

				_, err = w.Write([]byte(`{
					"access_token": "some-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
				Expect(err).ToNot(HaveOccurred())
			}
		}))
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/some/path" {
				authHeader = req.Header.Get("Authorization")

				w.WriteHeader(http.StatusNoContent)
			}
		}))
		accessURL = server.URL + "/some/path"
	})

	Describe("Do", func() {
		It("makes a request with client credentials", func() {
			client := NewOAuthClient(oauthServer.URL, "client_id", "client_secret", time.Duration(30)*time.Second, http.DefaultClient)

			req, err := http.NewRequest("GET", accessURL, strings.NewReader("request-body"))
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			Expect(authHeader).To(Equal("Bearer some-token"))

			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(receivedRequest)))
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/oauth/token"))

			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("client_id"))
			Expect(password).To(Equal("client_secret"))

			err = req.ParseForm()
			Expect(err).NotTo(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"grant_type": []string{"client_credentials"},
			}))
		})

		Context("when the target url is empty", func() {
			It("returns an error", func() {
				client := NewOAuthClient("", "", "", time.Duration(30)*time.Second, http.DefaultClient)

				req, err := http.NewRequest("GET", accessURL, strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				_, err = client.Do(req)
				Expect(err).To(MatchError(ContainSubstring("")))
			})
		})
	})
})
