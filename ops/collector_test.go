package ops_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	. "github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/ops/opsfakes"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
)

var _ = Describe("Collector", func() {
	var (
		dataCollector *opsfakes.FakeDataCollector
		tarWriter     *opsfakes.FakeTarWriter
		collector     CollectExecutor
	)

	BeforeEach(func() {
		dataCollector = new(opsfakes.FakeDataCollector)
		tarWriter = new(opsfakes.FakeTarWriter)

		collector = NewCollector(dataCollector, tarWriter)
	})

	It("collects data and writes it", func() {
		expectedD1Contents := "d1-content"
		md5SumD1 := md5.Sum([]byte(expectedD1Contents))
		d1ContentMd5 := base64.StdEncoding.EncodeToString(md5SumD1[:])
		expectedD2Contents := "d2-content"
		md5SumD2 := md5.Sum([]byte(expectedD2Contents))
		d2ContentMd5 := base64.StdEncoding.EncodeToString(md5SumD2[:])
		d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
		d2 := opsmanager.NewData(strings.NewReader(expectedD2Contents), "d2", "better-kind")
		dataToWrite := []opsmanager.Data{d1, d2}
		dataCollector.CollectReturns(dataToWrite, nil)

		envType := "most-production"
		err := collector.Collect(envType)
		Expect(err).NotTo(HaveOccurred())

		Expect(tarWriter.AddFileCallCount()).To(Equal(3))

		d1Contents, d1Name := tarWriter.AddFileArgsForCall(0)
		d2Contents, d2Name := tarWriter.AddFileArgsForCall(1)
		metadataContents, metadataName := tarWriter.AddFileArgsForCall(2)
		Expect(string(d1Contents)).To(Equal(expectedD1Contents))
		Expect(d1Name).To(Equal(d1.Name()))
		Expect(string(d2Contents)).To(Equal(expectedD2Contents))
		Expect(d2Name).To(Equal(d2.Name()))

		Expect(metadataName).To(Equal(MetadataFileName))
		var metadata Metadata
		Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
		Expect(metadata.EnvType).To(Equal(envType))
		Expect(metadata.FileDigests).To(ConsistOf(
			FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type()},
			FileDigest{Name: d2.Name(), MimeType: d2.MimeType(), MD5Checksum: d2ContentMd5, ProductType: d2.Type()},
		))
		_, err = uuid.FromString(metadata.CollectionId)
		Expect(err).NotTo(HaveOccurred())
		collectedAtTime, err := time.Parse(time.RFC3339, metadata.CollectedAt)
		Expect(err).NotTo(HaveOccurred())
		Expect(collectedAtTime.Location()).To(Equal(time.UTC))
		Expect(collectedAtTime).To(BeTemporally("~", time.Now(), time.Minute))

		Expect(tarWriter.CloseCallCount()).To(Equal(1))
	})

	It("returns an error when the collection errors", func() {
		dataCollector.CollectReturns([]opsmanager.Data{}, errors.New("collecting is hard"))

		err := collector.Collect("")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(CollectFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
	})

	It("returns an error when reading the data content fails", func() {
		failingReader := new(opsfakes.FakeReader)
		failingReader.ReadReturns(0, errors.New("reading is hard"))
		failingData := opsmanager.NewData(failingReader, "d1", "best-kind")
		dataCollector.CollectReturns([]opsmanager.Data{failingData}, nil)

		err := collector.Collect("")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("reading is hard")))
	})

	It("returns an error when adding data to the tar file fails", func() {
		data := opsmanager.NewData(strings.NewReader(""), "d1", "best-kind")
		dataCollector.CollectReturns([]opsmanager.Data{data}, nil)
		tarWriter.AddFileReturnsOnCall(0, errors.New("tarring is hard"))

		err := collector.Collect("")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when adding the metadata to the tar file fails", func() {
		tarWriter.AddFileReturns(errors.New("tarring is hard"))

		err := collector.Collect("")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})
})

//go:generate counterfeiter . reader
type reader interface {
	Read(p []byte) (n int, err error)
}
