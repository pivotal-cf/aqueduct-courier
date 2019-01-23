package operations_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/credhub"

	. "github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/aqueduct-utils/data"
	uuid "github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/operations/operationsfakes"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
)

var _ = Describe("Collector", func() {
	var (
		omDataCollector      *operationsfakes.FakeOmDataCollector
		credhubDataCollector *operationsfakes.FakeCredhubDataCollector
		tarWriter            *operationsfakes.FakeTarWriter
		collector            CollectExecutor
	)

	BeforeEach(func() {
		omDataCollector = new(operationsfakes.FakeOmDataCollector)
		credhubDataCollector = new(operationsfakes.FakeCredhubDataCollector)
		tarWriter = new(operationsfakes.FakeTarWriter)

		collector = NewCollector(omDataCollector, credhubDataCollector, tarWriter)
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
		omDataCollector.CollectReturns(dataToWrite, nil)

		expectedCHContents := "ch-content"
		md5SumCH := md5.Sum([]byte(expectedCHContents))
		chContentMd5 := base64.StdEncoding.EncodeToString(md5SumCH[:])
		chData := credhub.NewData(strings.NewReader(expectedCHContents))
		credhubDataCollector.CollectReturns(chData, nil)

		collectorVersion := "0.0.1-version"
		envType := "most-production"
		err := collector.Collect(envType, collectorVersion)
		Expect(err).NotTo(HaveOccurred())

		Expect(tarWriter.AddFileCallCount()).To(Equal(4))

		d1Contents, d1Name := tarWriter.AddFileArgsForCall(0)
		d2Contents, d2Name := tarWriter.AddFileArgsForCall(1)
		chContents, chName := tarWriter.AddFileArgsForCall(2)
		metadataContents, metadataName := tarWriter.AddFileArgsForCall(3)
		Expect(string(d1Contents)).To(Equal(expectedD1Contents))
		Expect(d1Name).To(Equal(d1.Name()))
		Expect(string(d2Contents)).To(Equal(expectedD2Contents))
		Expect(d2Name).To(Equal(d2.Name()))
		Expect(string(chContents)).To(Equal(expectedCHContents))
		Expect(chName).To(Equal(chData.Name()))

		Expect(metadataName).To(Equal(data.MetadataFileName))
		var metadata data.Metadata
		Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
		Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
		Expect(metadata.EnvType).To(Equal(envType))
		Expect(metadata.FileDigests).To(ConsistOf(
			data.FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type(), DataType: d1.DataType()},
			data.FileDigest{Name: d2.Name(), MimeType: d2.MimeType(), MD5Checksum: d2ContentMd5, ProductType: d2.Type(), DataType: d2.DataType()},
			data.FileDigest{Name: chData.Name(), MimeType: chData.MimeType(), MD5Checksum: chContentMd5, ProductType: chData.Type(), DataType: chData.DataType()},
		))
		_, err = uuid.FromString(metadata.CollectionId)
		Expect(err).NotTo(HaveOccurred())
		collectedAtTime, err := time.Parse(time.RFC3339, metadata.CollectedAt)
		Expect(err).NotTo(HaveOccurred())
		Expect(collectedAtTime.Location()).To(Equal(time.UTC))
		Expect(collectedAtTime).To(BeTemporally("~", time.Now(), time.Minute))

		Expect(tarWriter.CloseCallCount()).To(Equal(1))
	})

	It("returns an error when the ops manager collection errors", func() {
		omDataCollector.CollectReturns([]opsmanager.Data{}, errors.New("collecting is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(OpsManagerCollectFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
	})

	It("returns an error when the credhub collection errors", func() {
		credhubDataCollector.CollectReturns(credhub.Data{}, errors.New("collecting is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(CredhubCollectFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
	})

	It("returns an error when reading the ops manager data content fails", func() {
		failingReader := new(operationsfakes.FakeReader)
		failingReader.ReadReturns(0, errors.New("reading is hard"))
		failingData := opsmanager.NewData(failingReader, "d1", "best-kind")
		omDataCollector.CollectReturns([]opsmanager.Data{failingData}, nil)

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("reading is hard")))
	})

	It("returns an error when reading the credhub data content fails", func() {
		failingReader := new(operationsfakes.FakeReader)
		failingReader.ReadReturns(0, errors.New("reading is hard"))
		failingData := credhub.NewData(failingReader)
		credhubDataCollector.CollectReturns(failingData, nil)

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("reading is hard")))
	})

	It("returns an error when adding ops manager data to the tar file fails", func() {
		data := opsmanager.NewData(strings.NewReader(""), "d1", "best-kind")
		omDataCollector.CollectReturns([]opsmanager.Data{data}, nil)
		tarWriter.AddFileReturnsOnCall(0, errors.New("tarring is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when adding credhub data to the tar file fails", func() {
		data := credhub.NewData(strings.NewReader(""))
		credhubDataCollector.CollectReturns(data, nil)
		tarWriter.AddFileReturnsOnCall(0, errors.New("tarring is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when adding the metadata to the tar file fails", func() {
		tarWriter.AddFileReturns(errors.New("tarring is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})
})

//go:generate counterfeiter . reader
type reader interface {
	io.Reader
}
