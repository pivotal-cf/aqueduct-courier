package ops_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/file/filefakes"
	. "github.com/pivotal-cf/aqueduct-courier/ops"
)

var _ = Describe("Sender", func() {
	var (
		dataLoader    *ghttp.Server
		sender        SendExecutor
		dataDirectory string
		fillerData    *filefakes.FakeData
	)

	BeforeEach(func() {
		var err error
		dataLoader = ghttp.NewServer()
		sender = SendExecutor{}

		dataDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		fillerData = new(filefakes.FakeData)
		fillerData.NameReturns("filler-name")
		fillerData.MimeTypeReturns("not-a-real-mime")
		fillerData.ContentReturns(strings.NewReader(""))
		writer := file.NewWriter("")
		writer.Write(fillerData, dataDirectory)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dataDirectory)).To(Succeed())

		dataLoader.Close()
	})

	It("posts to the data loader with the file as content and the file metadata", func() {
		//Setup
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		envType := "any-env-type"
		writer := file.NewWriter(envType)

		d1 := new(filefakes.FakeData)
		d1.NameReturns("data-file1")
		d1.ContentReturns(strings.NewReader("data-file1-contents"))
		d1.MimeTypeReturns("data-file1-mimetype")

		d2 := new(filefakes.FakeData)
		d2.NameReturns("data-file2")
		d2.ContentReturns(strings.NewReader("data-file2-contents"))
		d2.MimeTypeReturns("data-file2-mimetype")

		err = writer.Write(d1, dir)
		Expect(err).NotTo(HaveOccurred())
		err = writer.Write(d2, dir)
		Expect(err).NotTo(HaveOccurred())

		mFile, err := ioutil.ReadFile(filepath.Join(dir, file.MetadataFileName))
		Expect(err).NotTo(HaveOccurred())

		var fileMetadata file.Metadata
		err = json.Unmarshal(mFile, &fileMetadata)
		Expect(err).NotTo(HaveOccurred())

		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				f, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				contents, err := ioutil.ReadAll(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(fileHeaders.Filename + "-contents"))

				metadataStr := req.FormValue("metadata")
				var metadata map[string]string
				sum := md5.Sum([]byte(fileHeaders.Filename + "-contents"))
				checksum := base64.StdEncoding.EncodeToString(sum[:])
				Expect(json.Unmarshal([]byte(metadataStr), &metadata)).To(Succeed())
				Expect(metadata["filename"]).To(Equal(fileHeaders.Filename))
				Expect(metadata["fileContentType"]).To(Equal(fileHeaders.Filename + "-mimetype"))
				Expect(metadata["fileMd5Checksum"]).To(Equal(checksum))
				Expect(metadata["collectedAt"]).To(Equal(fileMetadata.CollectedAt))
				Expect(metadata["envType"]).To(Equal(envType))
			},
			ghttp.RespondWith(http.StatusCreated, ""),
		))

		//Test
		Expect(sender.Send(dir, dataLoader.URL(), "some-key")).To(Succeed())
		reqs := dataLoader.ReceivedRequests()
		Expect(len(reqs)).To(Equal(2))

		verifyFileSentInRequest(d1.Name(), reqs[0])
		verifyFileSentInRequest(d2.Name(), reqs[1])

		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("posts to the data loader with the correct API key in the header", func() {
		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.VerifyHeader(http.Header{
				"Authorization": []string{"Token some-key"},
			}),
			ghttp.RespondWith(http.StatusCreated, ""),
		))
		Expect(sender.Send(dataDirectory, dataLoader.URL(), "some-key")).To(Succeed())
	})

	It("does not post files/data not in the metadata file", func() {
		Expect(ioutil.WriteFile(filepath.Join(dataDirectory, "file-not-to-send"), []byte(""), 0644)).To(Succeed())

		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.RespondWith(http.StatusCreated, ""),
			func(w http.ResponseWriter, req *http.Request) {
				_, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				Expect(fileHeaders.Filename).NotTo(Equal("file-not-to-send"))
			},
		))

		Expect(sender.Send(dataDirectory, dataLoader.URL(), "some-key")).To(Succeed())
		reqs := dataLoader.ReceivedRequests()
		Expect(len(reqs)).To(Equal(1))
	})

	It("fails if the metadata file cannot be unmarshalled", func() {
		Expect(ioutil.WriteFile(filepath.Join(dataDirectory, file.MetadataFileName), []byte("][,.dsf..invalid"), 0644)).To(Succeed())

		err := sender.Send(dataDirectory, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(InvalidMetadataFileErrorFormat, filepath.Join(dataDirectory, file.MetadataFileName)))))
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(dataDirectory, "127.0.0.1:a", "some-key")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		err := sender.Send(dataDirectory, "http://127.0.0.1:999999", "some-key")
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		dataLoader.AppendHandlers(
			ghttp.RespondWith(http.StatusUnauthorized, ""),
		)

		err := sender.Send(dataDirectory, dataLoader.URL(), "invalid-key")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedResponseCodeFormat, http.StatusUnauthorized)))
	})

	It("errors when the metadata file does not exist", func() {
		err := sender.Send("/does/not/exist", dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(ReadMetadataFileErrorFormat, filepath.Clean("/does/not/exist/aqueduct_metadata")))))
	})

	It("errors when the data files are not sibling to the metadata file", func() {
		Expect(os.Remove(filepath.Join(dataDirectory, fillerData.Name()))).To(Succeed())
		err := sender.Send(dataDirectory, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})
})

func verifyFileSentInRequest(filename string, req *http.Request) {
	_, fileHeader, err := req.FormFile("data")
	Expect(err).NotTo(HaveOccurred())
	Expect(fileHeader.Filename).To(Equal(filename))
}
