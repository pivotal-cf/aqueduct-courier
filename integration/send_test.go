package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/ops"
)

var _ = Describe("Send", func() {
	var (
		binaryPath            string
		dataLoader            *ghttp.Server
		tempDir               string
		sourceDataTarFilePath string
		validApiKey           = "best-key"
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

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		sourceDataTarFilePath = generateValidDataTarFile(tempDir)
	})

	AfterEach(func() {
		dataLoader.Close()
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Context("success", func() {
		BeforeEach(func() {
			dataLoader.RouteToHandler(http.MethodPost, ops.PostPath, ghttp.CombineHandlers(
				ghttp.VerifyHeader(http.Header{
					"Authorization": []string{fmt.Sprintf("Token %s", validApiKey)},
				}),
				ghttp.RespondWith(http.StatusCreated, ""),
			))
		})

		It("sends data to the configured endpoint with flag configuration", func() {
			command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath, "--api-key="+validApiKey)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(2))
		})

		It("sends data to the configured endpoint with api key as an env variable", func() {
			command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath)
			command.Env = append(os.Environ(), fmt.Sprintf("%s=%s", cmd.ApiKeyKey, validApiKey))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(2))
		})
	})

	It("exits non-zero when sending to pivotal fails", func() {
		dataLoader.RouteToHandler(http.MethodPost, ops.PostPath, ghttp.RespondWith(http.StatusUnauthorized, ""))

		command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath, "--api-key=incorrect-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(cmd.SendFailureMessage))
		Expect(session.Err).NotTo(gbytes.Say("Usage:"))
	})

	It("fails if the path flag has not been set", func() {
		command := exec.Command(binaryPath, "send", "--api-key="+validApiKey)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.DataTarFilePathFlag)))
		Expect(session.Err).To(gbytes.Say("Usage:"))
	})

	It("fails if the api-key flag has not been set", func() {
		command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.ApiKeyFlag)))
		Expect(session.Err).To(gbytes.Say("Usage:"))
	})
})

func generateValidDataTarFile(destinationDir string) string {
	tarFilePath := filepath.Join(destinationDir, "some-foundation-data")

	writer, err := file.NewTarWriter(tarFilePath)
	Expect(err).NotTo(HaveOccurred())

	Expect(writer.AddFile([]byte(""), "file1")).To(Succeed())
	Expect(writer.AddFile([]byte(""), "file2")).To(Succeed())

	var metadata ops.Metadata
	metadata.FileDigests = []ops.FileDigest{
		{Name: "file1"},
		{Name: "file2"},
	}
	metadataContents, err := json.Marshal(metadata)
	Expect(err).NotTo(HaveOccurred())
	Expect(writer.AddFile(metadataContents, ops.MetadataFileName)).To(Succeed())

	return tarFilePath
}
