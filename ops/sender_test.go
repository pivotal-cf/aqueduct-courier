package ops_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-courier/ops/opsfakes"
	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pkg/errors"
)

var _ = Describe("Sender", func() {
	var (
		dataLoader *ghttp.Server
		tarReader  *opsfakes.FakeTarReader
		validator  *opsfakes.FakeValidator
		metadata   data.Metadata
		tmpFile    *os.File
		tarContent string
		sender     SendExecutor
	)

	BeforeEach(func() {
		dataLoader = ghttp.NewServer()
		sender = SendExecutor{}

		tarReader = new(opsfakes.FakeTarReader)
		validator = new(opsfakes.FakeValidator)

		metadata = data.Metadata{
			CollectedAt:  "collected-at",
			CollectionId: "collection-id",
			EnvType:      "some-env-type",
			FileDigests: []data.FileDigest{
				{Name: "file1", MD5Checksum: "file1-md5"},
				{Name: "file2", MD5Checksum: "file2-md5"},
			},
		}
		metadataContents, err := json.Marshal(metadata)
		Expect(err).NotTo(HaveOccurred())
		tarReader.FileMd5sReturns(
			map[string]string{
				"file1": "file1-md5",
				"file2": "file2-md5",
			},
			nil,
		)

		tmpFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		tarContent = "tar-content"
		_, err = tmpFile.Write([]byte(tarContent))
		Expect(err).NotTo(HaveOccurred())
		Expect(tmpFile.Close()).To(Succeed())

		tarReader.ReadFileStub = func(fileName string) ([]byte, error) {
			if fileName == data.MetadataFileName {
				return metadataContents, nil
			}

			return []byte{}, errors.New("unexpected file requested")
		}
	})

	AfterEach(func() {
		dataLoader.Close()
		Expect(os.RemoveAll(tmpFile.Name())).To(Succeed())
	})

	It("posts to the data loader with the file as content and the file metadata", func() {
		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				f, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				contents, err := ioutil.ReadAll(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(tarContent))

				metadataStr := req.FormValue("metadata")
				var metadataMap map[string]string
				Expect(json.Unmarshal([]byte(metadataStr), &metadataMap)).To(Succeed())

				Expect(metadataMap["filename"]).To(Equal(fileHeaders.Filename))
				Expect(metadataMap["envType"]).To(Equal(metadata.EnvType))
				Expect(metadataMap["collectedAt"]).To(Equal(metadata.CollectedAt))
				Expect(metadataMap["collectionId"]).To(Equal(metadata.CollectionId))
				Expect(metadataMap["fileContentType"]).To(Equal(TarMimeType))

				md5Sum := md5.Sum([]byte(tarContent))
				Expect(metadataMap["fileMd5Checksum"]).To(Equal(base64.StdEncoding.EncodeToString(md5Sum[:])))
			},
			ghttp.RespondWith(http.StatusCreated, ""),
		))

		Expect(sender.Send(tarReader, validator, tmpFile.Name(), dataLoader.URL(), "some-key")).To(Succeed())

		reqs := dataLoader.ReceivedRequests()
		Expect(len(reqs)).To(Equal(1))
	})

	It("posts to the data loader with the correct API key in the header", func() {
		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.VerifyHeader(http.Header{
				"Authorization": []string{"Token some-key"},
			}),
			ghttp.RespondWith(http.StatusCreated, ""),
		))
		Expect(sender.Send(tarReader, validator, tmpFile.Name(), dataLoader.URL(), "some-key")).To(Succeed())
	})

	It("errors when validation fails", func() {
		validator.ValidateReturns(errors.New("totally invalid tar"))
		err := sender.Send(tarReader, validator, "path/to/file", dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(FileValidationFailedMessageFormat, "path/to/file"))))
	})

	It("fails if the metadata file cannot be unmarshalled", func() {
		tarReader.ReadFileReturns([]byte("some-bad-metadata"), nil)

		err := sender.Send(tarReader, validator, tmpFile.Name(), dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(InvalidMetadataFileError)))
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(tarReader, validator, tmpFile.Name(), "127.0.0.1:a", "some-key")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		err := sender.Send(tarReader, validator, tmpFile.Name(),"http://127.0.0.1:999999", "some-key")
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		dataLoader.AppendHandlers(
			ghttp.RespondWith(http.StatusUnauthorized, ""),
		)

		err := sender.Send(tarReader, validator, tmpFile.Name(), dataLoader.URL(), "invalid-key")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedResponseCodeFormat, http.StatusUnauthorized)))
	})

	It("when the tarFile does not exist", func() {
		err := sender.Send(tarReader, validator, "path/to/not/the/tarFile", dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(ReadDataFileError)))
	})
})
