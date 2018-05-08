package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/file/filefakes"
	"github.com/pivotal-cf/aqueduct-courier/ops"
)

var _ = Describe("Send", func() {
	var (
		binaryPath string
		dataLoader *ghttp.Server
	)

	BeforeEach(func() {
		dataLoader = ghttp.NewServer()

		var err error
		binaryPath, err = gexec.Build(
			"github.com/pivotal-cf/aqueduct-courier",
			"-ldflags",
			fmt.Sprintf("-X github.com/pivotal-cf/aqueduct-courier/cmd.dataLoaderURL=%s", dataLoader.URL()),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		dataLoader.Close()
	})

	Context("success", func() {
		var (
			dir    string
			apiKey = "best-key"
		)

		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			d1 := new(filefakes.FakeData)
			d1.NameReturns("data-file1")
			d1.ContentReturns(strings.NewReader("data-file1-contents"))
			d1.MimeTypeReturns("data-file1-mimetype")
			writer := &file.Writer{}
			err = writer.Write(d1, dir, "")
			Expect(err).NotTo(HaveOccurred())

			dataLoader.RouteToHandler(http.MethodPost, ops.PostPath, ghttp.CombineHandlers(
				ghttp.VerifyHeader(http.Header{
					"Authorization": []string{fmt.Sprintf("Token %s", apiKey)},
				}),
				ghttp.RespondWith(http.StatusCreated, ""),
			))
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("sends data to the configured endpoint with flag configuration", func() {
			command := exec.Command(binaryPath, "send", "--path="+dir, fmt.Sprintf("--api-key=%s", apiKey))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(1))
		})

		It("sends data to the configured endpoint with api key as an env variable", func() {
			command := exec.Command(binaryPath, "send", "--path="+dir)
			command.Env = append(os.Environ(), fmt.Sprintf("%s=%s", cmd.ApiKeyKey, apiKey))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(1))
		})
	})

	It("exits non-zero when sending to pivotal fails", func() {
		command := exec.Command(binaryPath, "send", "--path=/path/to/data", "--api-key=incorrect-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(cmd.SendFailureMessage))
	})

	It("fails if the path flag has not been set", func() {
		command := exec.Command(binaryPath, "send", "--api-key=best-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.DirectoryPathFlag)))
	})

	It("fails if the api-key flag has not been set", func() {
		command := exec.Command(binaryPath, "send", "--path=/path/to/data")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.ApiKeyFlag)))
	})
})
