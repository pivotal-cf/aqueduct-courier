package ops_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-courier/ops/opsfakes"
	"github.com/pkg/errors"
)

var _ = Describe("Sender", func() {
	var (
		dataLoader *ghttp.Server
		tarReader  *opsfakes.FakeTarReader

		sender SendExecutor
	)

	BeforeEach(func() {
		dataLoader = ghttp.NewServer()
		sender = SendExecutor{}

		tarReader = new(opsfakes.FakeTarReader)

		metadata := Metadata{
			CollectedAt:  "collected-at",
			CollectionId: "collection-id",
			EnvType:      "some-env-type",
			FileDigests: []FileDigest{
				{Name: "file1", MD5Checksum: "file1-checksum", MimeType: "file1-mimetype"},
			},
		}
		metadataContents, err := json.Marshal(metadata)
		Expect(err).NotTo(HaveOccurred())

		tarReader.ReadFileStub = func(fileName string) ([]byte, error) {
			if fileName == MetadataFileName {
				return metadataContents, nil
			}

			if fileName == metadata.FileDigests[0].Name {
				return []byte("file1-contents"), nil
			}

			return []byte{}, errors.New("unexpected file requested")
		}

	})

	AfterEach(func() {
		dataLoader.Close()
	})

	It("posts to the data loader with the file as content and the file metadata", func() {
		metadata := Metadata{
			CollectedAt:  "collected-at",
			CollectionId: "collection-id",
			EnvType:      "some-env-type",
			FileDigests: []FileDigest{
				{Name: "file1", MD5Checksum: "file1-checksum", MimeType: "file1-mimetype"},
				{Name: "file2", MD5Checksum: "file2-checksum", MimeType: "file2-mimetype"},
			},
		}
		metadataContents, err := json.Marshal(metadata)
		Expect(err).NotTo(HaveOccurred())

		tarReader.ReadFileStub = func(fileName string) ([]byte, error) {
			if fileName == MetadataFileName {
				return metadataContents, nil
			}

			if fileName == metadata.FileDigests[0].Name {
				return []byte("file1-contents"), nil
			}

			if fileName == metadata.FileDigests[1].Name {
				return []byte("file2-contents"), nil
			}

			return []byte{}, errors.New("unexpected file requested")
		}

		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			func(w http.ResponseWriter, req *http.Request) {
				f, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				contents, err := ioutil.ReadAll(f)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(fileHeaders.Filename + "-contents"))

				metadataStr := req.FormValue("metadata")
				var metadataMap map[string]string
				Expect(json.Unmarshal([]byte(metadataStr), &metadataMap)).To(Succeed())

				Expect(metadataMap["filename"]).To(Equal(fileHeaders.Filename))
				Expect(metadataMap["fileContentType"]).To(Equal(fileHeaders.Filename + "-mimetype"))
				Expect(metadataMap["fileMd5Checksum"]).To(Equal(fileHeaders.Filename + "-checksum"))
				Expect(metadataMap["collectedAt"]).To(Equal(metadata.CollectedAt))
				Expect(metadataMap["envType"]).To(Equal(metadata.EnvType))
				Expect(metadataMap["collectionId"]).To(Equal(metadata.CollectionId))
			},
			ghttp.RespondWith(http.StatusCreated, ""),
		))

		Expect(sender.Send(tarReader, dataLoader.URL(), "some-key")).To(Succeed())

		reqs := dataLoader.ReceivedRequests()
		Expect(len(reqs)).To(Equal(2))

		verifyFileSentInRequest(metadata.FileDigests[0].Name, reqs[0])
		verifyFileSentInRequest(metadata.FileDigests[1].Name, reqs[1])
	})

	It("posts to the data loader with the correct API key in the header", func() {
		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.VerifyHeader(http.Header{
				"Authorization": []string{"Token some-key"},
			}),
			ghttp.RespondWith(http.StatusCreated, ""),
		))
		Expect(sender.Send(tarReader, dataLoader.URL(), "some-key")).To(Succeed())
	})

	It("errors when the metadata file does not exist", func() {
		tarReader.ReadFileReturns([]byte{}, errors.New("can't find the metadata file"))
		err := sender.Send(tarReader, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(ReadMetadataFileError)))
	})

	It("fails if the metadata file cannot be unmarshalled", func() {
		tarReader.ReadFileReturns([]byte("some-bad-metadata"), nil)

		err := sender.Send(tarReader, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(InvalidMetadataFileError)))
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(tarReader, "127.0.0.1:a", "some-key")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		err := sender.Send(tarReader, "http://127.0.0.1:999999", "some-key")
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		dataLoader.AppendHandlers(
			ghttp.RespondWith(http.StatusUnauthorized, ""),
		)

		err := sender.Send(tarReader, dataLoader.URL(), "invalid-key")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedResponseCodeFormat, http.StatusUnauthorized)))
	})

	It("errors when reading the data files fail", func() {
		metadata := Metadata{
			FileDigests: []FileDigest{
				{Name: "file1"},
			},
		}
		metadataContents, err := json.Marshal(metadata)
		Expect(err).NotTo(HaveOccurred())
		tarReader.ReadFileReturnsOnCall(0, metadataContents, nil)
		tarReader.ReadFileReturnsOnCall(1, []byte{}, errors.New("failed to read data file"))

		err = sender.Send(tarReader, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(ReadDataFileError)))
	})
})

func verifyFileSentInRequest(filename string, req *http.Request) {
	_, fileHeader, err := req.FormFile("data")
	Expect(err).NotTo(HaveOccurred())
	Expect(fileHeader.Filename).To(Equal(filename))
}
