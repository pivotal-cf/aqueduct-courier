package integration

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-utils/data"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-utils/file"
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
			fmt.Sprintf("-X github.com/pivotal-cf/aqueduct-courier/cmd.dataLoaderURL=%s -X github.com/pivotal-cf/aqueduct-courier/cmd.version=%s", dataLoader.URL(), testVersion),
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
				verifyVersion(),
			))
		})

		It("sends data to the configured endpoint with flag configuration", func() {
			command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath, "--api-key="+validApiKey)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(1))
			Expect(session.Out).To(gbytes.Say(fmt.Sprintf("Sending %s to Pivotal at %s\n", escapeWindowsPathRegex(sourceDataTarFilePath), dataLoader.URL())))
			Expect(session.Out).To(gbytes.Say("Success!\n"))
		})

		It("sends data to the configured endpoint with api key as an env variable", func() {
			command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath)
			command.Env = append(os.Environ(), fmt.Sprintf("%s=%s", cmd.ApiKeyKey, validApiKey))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(len(dataLoader.ReceivedRequests())).To(Equal(1))
			Expect(session.Out).To(gbytes.Say(fmt.Sprintf("Sending %s to Pivotal at %s\n", escapeWindowsPathRegex(sourceDataTarFilePath), dataLoader.URL())))
			Expect(session.Out).To(gbytes.Say("Success!\n"))
		})
	})

	It("exits non-zero when sending to pivotal fails", func() {
		dataLoader.RouteToHandler(http.MethodPost, ops.PostPath, ghttp.RespondWith(http.StatusUnauthorized, ""))

		command := exec.Command(binaryPath, "send", "--path="+sourceDataTarFilePath, "--api-key=incorrect-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(cmd.SendFailureMessage))
		Expect(session.Err).NotTo(gbytes.Say("Usage:"))
	})

	It("fails if required flags have not been set", func() {
		command := exec.Command(binaryPath, "send")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		requiredFlags := []string{"--" + cmd.DataTarFilePathFlag, "--" + cmd.ApiKeyFlag}
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, strings.Join(requiredFlags, ", "))))
		Expect(session.Err).To(gbytes.Say("Usage:"))
	})
})

func generateValidDataTarFile(destinationDir string) string {
	tarFilePath := filepath.Join(destinationDir, "some-foundation-data")

	writer, err := file.NewTarWriter(tarFilePath)
	Expect(err).NotTo(HaveOccurred())
	defer writer.Close()

	Expect(writer.AddFile([]byte{}, "file1")).To(Succeed())
	Expect(writer.AddFile([]byte{}, "file2")).To(Succeed())
	sum := md5.Sum([]byte{})
	emptyFileChecksum := base64.StdEncoding.EncodeToString(sum[:])

	var metadata data.Metadata
	metadata.FileDigests = []data.FileDigest{
		{Name: "file1", MD5Checksum: emptyFileChecksum},
		{Name: "file2", MD5Checksum: emptyFileChecksum},
	}
	metadataContents, err := json.Marshal(metadata)
	Expect(err).NotTo(HaveOccurred())
	Expect(writer.AddFile(metadataContents, data.MetadataFileName)).To(Succeed())

	return tarFilePath
}

func verifyVersion() http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		metadataStr := req.FormValue("metadata")
		var metadataMap map[string]interface{}
		Expect(json.Unmarshal([]byte(metadataStr), &metadataMap)).To(Succeed())

		customMetadataMap, ok := metadataMap["customMetadata"].(map[string]interface{})
		Expect(ok).To(BeTrue())

		Expect(customMetadataMap["SenderVersion"]).To(Equal(testVersion))
	}
}
