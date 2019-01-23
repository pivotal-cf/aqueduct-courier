package operations_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pivotal-cf/aqueduct-utils/urd"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/aqueduct-courier/operations/operationsfakes"
	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pkg/errors"
)

var _ = Describe("Sender", func() {
	var (
		client     *operationsfakes.FakeHttpClient
		tarReader  *operationsfakes.FakeTarReader
		validator  *operationsfakes.FakeValidator
		metadata   data.Metadata
		tmpFile    *os.File
		tarContent string
		sender     SendExecutor
	)

	BeforeEach(func() {
		sender = SendExecutor{}

		client = new(operationsfakes.FakeHttpClient)
		tarReader = new(operationsfakes.FakeTarReader)
		validator = new(operationsfakes.FakeValidator)

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
		emptyBody := ioutil.NopCloser(strings.NewReader(""))
		client.DoReturns(&http.Response{StatusCode: http.StatusCreated, Body: emptyBody}, nil)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpFile.Name())).To(Succeed())
	})

	It("posts to the data loader with the file as content and the file metadata", func() {
		senderVersion := "best-sender-version"
		Expect(sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "some-key", senderVersion)).To(Succeed(), "")

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)
		Expect(req.URL.String()).To(Equal(fmt.Sprintf("http://example.com%s", PostPath)))
		f, fileHeaders, err := req.FormFile("data")
		Expect(err).ToNot(HaveOccurred())
		contents, err := ioutil.ReadAll(f)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(contents)).To(Equal(tarContent))

		metadataStr := req.FormValue("metadata")
		var urdMetadata urd.Metadata
		Expect(json.Unmarshal([]byte(metadataStr), &urdMetadata)).To(Succeed())

		md5Sum := md5.Sum([]byte(tarContent))
		Expect(urdMetadata).To(Equal(urd.Metadata{
			Filename:        fileHeaders.Filename,
			FileContentType: TarMimeType,
			FileMD5Checksum: base64.StdEncoding.EncodeToString(md5Sum[:]),
			CollectedAt:     metadata.CollectedAt,
			CustomMetadata: map[string]interface{}{
				"senderVersion": senderVersion,
			},
		}))
	})

	It("posts to the data loader with the correct API key in the header", func() {
		Expect(sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "some-key", "")).To(Succeed())
		req := client.DoArgsForCall(0)
		Expect(req.Header.Get("Authorization")).To(Equal("Token some-key"))
	})

	It("errors when validation fails", func() {
		validator.ValidateReturns(errors.New("totally invalid tar"))
		err := sender.Send(client, tarReader, validator, "path/to/file", "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(FileValidationFailedMessageFormat, "path/to/file"))))
	})

	It("fails if the metadata file cannot be unmarshalled", func() {
		tarReader.ReadFileReturns([]byte("some-bad-metadata"), nil)

		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(InvalidMetadataFileError)))
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "127.0.0.1:a", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		client.DoReturns(nil, errors.New("doing requests is hard"))
		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring("doing requests is hard")))
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		emptyBody := ioutil.NopCloser(strings.NewReader(""))
		client.DoReturns(&http.Response{StatusCode: http.StatusUnauthorized, Body: emptyBody}, nil)

		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(UnauthorizedErrorMessage))
	})

	It("errors if the error response cannot be read", func() {
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: ioutil.NopCloser(&badReader{})}, nil)
		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "unknown")))
	})

	It("errors if the error response cannot be read into the expected structure", func() {
		badBody := ioutil.NopCloser(strings.NewReader(`{not json`))
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: badBody}, nil)
		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "unknown")))
	})

	It("errors when the response code is not 201/401", func() {
		emptyBody := ioutil.NopCloser(strings.NewReader(`{"error": {"uuid": "error-uuid"}}`))
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: emptyBody}, nil)

		err := sender.Send(client, tarReader, validator, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "error-uuid")))
	})

	It("when the tarFile does not exist", func() {
		err := sender.Send(client, tarReader, validator, "path/to/not/the/tarFile", "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(ReadDataFileError)))
	})
})

type badReader struct{}

func (r *badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("reading is hard")
}
