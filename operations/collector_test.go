package operations_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/consumption"

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
		omDataCollector *operationsfakes.FakeOmDataCollector
		tarWriter       *operationsfakes.FakeTarWriter
		collector       CollectExecutor
	)

	BeforeEach(func() {
		omDataCollector = new(operationsfakes.FakeOmDataCollector)
		tarWriter = new(operationsfakes.FakeTarWriter)

		collector = NewCollector(omDataCollector, nil, nil, tarWriter)
	})

	It("collects opsmanager data and writes it", func() {
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

		collectorVersion := "0.0.1-version"
		envType := "most-production"

		err := collector.Collect(envType, collectorVersion)
		Expect(err).NotTo(HaveOccurred())

		Expect(tarWriter.AddFileCallCount()).To(Equal(3))

		expectedD1Path := filepath.Join(data.OpsManagerCollectorDataSetId, d1.Name())
		d1Contents, d1Path := tarWriter.AddFileArgsForCall(0)
		expectedD2Path := filepath.Join(data.OpsManagerCollectorDataSetId, d2.Name())
		d2Contents, d2Path := tarWriter.AddFileArgsForCall(1)

		Expect(string(d1Contents)).To(Equal(expectedD1Contents))
		Expect(d1Path).To(Equal(expectedD1Path))
		Expect(string(d2Contents)).To(Equal(expectedD2Contents))
		Expect(d2Path).To(Equal(expectedD2Path))

		expectedMetadataPath := filepath.Join(data.OpsManagerCollectorDataSetId, data.MetadataFileName)
		metadataContents, metadataPath := tarWriter.AddFileArgsForCall(2)
		Expect(metadataPath).To(Equal(expectedMetadataPath))

		var metadata data.Metadata
		Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
		Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
		Expect(metadata.EnvType).To(Equal(envType))
		Expect(metadata.FileDigests).To(ConsistOf(
			data.FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type(), DataType: d1.DataType()},
			data.FileDigest{Name: d2.Name(), MimeType: d2.MimeType(), MD5Checksum: d2ContentMd5, ProductType: d2.Type(), DataType: d2.DataType()},
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

	It("returns an error when adding ops manager data to the tar file fails", func() {
		data := opsmanager.NewData(strings.NewReader(""), "d1", "best-kind")
		omDataCollector.CollectReturns([]opsmanager.Data{data}, nil)
		tarWriter.AddFileReturnsOnCall(0, errors.New("tarring is hard"))

		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when adding the metadata to the tar file fails", func() {
		tarWriter.AddFileStub = func(contents []byte, filePath string) error {
			if filePath == filepath.Join(data.OpsManagerCollectorDataSetId, data.MetadataFileName) {
				return errors.New("tarring is hard")
			}
			return nil
		}
		err := collector.Collect("", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	Describe("credhub collection", func() {
		var (
			collectorWithCredhub CollectExecutor
			credhubDataCollector *operationsfakes.FakeCredhubDataCollector
		)

		BeforeEach(func() {
			credhubDataCollector = new(operationsfakes.FakeCredhubDataCollector)
			collectorWithCredhub = NewCollector(omDataCollector, credhubDataCollector, nil, tarWriter)
		})

		It("collects credhub data and writes it", func() {
			expectedD1Contents := "d1-content"
			md5SumD1 := md5.Sum([]byte(expectedD1Contents))
			d1ContentMd5 := base64.StdEncoding.EncodeToString(md5SumD1[:])
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			omDataCollector.CollectReturns(dataToWrite, nil)

			expectedCHContents := "ch-content"
			md5SumCH := md5.Sum([]byte(expectedCHContents))
			chContentMd5 := base64.StdEncoding.EncodeToString(md5SumCH[:])
			chData := credhub.NewData(strings.NewReader(expectedCHContents))
			credhubDataCollector.CollectReturns(chData, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"

			err := collectorWithCredhub.Collect(envType, collectorVersion)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(3))

			chContents, credhubDataPath := tarWriter.AddFileArgsForCall(1)
			Expect(string(chContents)).To(Equal(expectedCHContents))

			expectedCredhubDataPath := filepath.Join(data.OpsManagerCollectorDataSetId, chData.Name())
			Expect(credhubDataPath).To(Equal(expectedCredhubDataPath))

			expectedMetadataPath := filepath.Join(data.OpsManagerCollectorDataSetId, data.MetadataFileName)
			metadataContents, metadataPath := tarWriter.AddFileArgsForCall(2)

			Expect(metadataPath).To(Equal(expectedMetadataPath))
			var metadata data.Metadata
			Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
			Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
			Expect(metadata.EnvType).To(Equal(envType))
			Expect(metadata.FileDigests).To(ConsistOf(
				data.FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type(), DataType: d1.DataType()},
				data.FileDigest{Name: chData.Name(), MimeType: chData.MimeType(), MD5Checksum: chContentMd5, ProductType: chData.Type(), DataType: chData.DataType()},
			))

			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("returns an error when the credhub collection errors", func() {
			credhubDataCollector.CollectReturns(credhub.Data{}, errors.New("collecting is hard"))

			err := collectorWithCredhub.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(CredhubCollectFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
		})

		It("returns an error when reading the credhub data content fails", func() {
			failingReader := new(operationsfakes.FakeReader)
			failingReader.ReadReturns(0, errors.New("reading is hard"))
			failingData := credhub.NewData(failingReader)
			credhubDataCollector.CollectReturns(failingData, nil)

			err := collectorWithCredhub.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("reading is hard")))
		})

		It("returns an error when adding credhub data to the tar file fails", func() {
			credhubData := credhub.NewData(strings.NewReader(""))
			credhubDataCollector.CollectReturns(credhubData, nil)
			tarWriter.AddFileStub = func(content []byte, filePath string) error {
				if filePath == filepath.Join(data.OpsManagerCollectorDataSetId, credhubData.Name()) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithCredhub.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})
	})

	Describe("consumption collection", func() {
		var (
			collectorWithConsumption CollectExecutor
			consumptionDataCollector *operationsfakes.FakeConsumptionDataCollector
		)

		BeforeEach(func() {
			consumptionDataCollector = new(operationsfakes.FakeConsumptionDataCollector)
			collectorWithConsumption = NewCollector(omDataCollector, nil, consumptionDataCollector, tarWriter)
		})

		It("collects consumption data and writes it", func() {
			expectedD1Contents := "d1-content"
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			omDataCollector.CollectReturns(dataToWrite, nil)

			expectedConsumptionContents := "consumption-content"
			md5SumCH := md5.Sum([]byte(expectedConsumptionContents))
			chContentMd5 := base64.StdEncoding.EncodeToString(md5SumCH[:])
			consumptionData := consumption.NewData(strings.NewReader(expectedConsumptionContents), "app-instances")
			consumptionDataCollector.CollectReturns(consumptionData, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"

			err := collectorWithConsumption.Collect(envType, collectorVersion)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(4))

			expectedConsumptionDataPath := filepath.Join(data.ConsumptionCollectorDataSetId, consumptionData.Name())
			consumptionContents, consumptionDataPath := tarWriter.AddFileArgsForCall(2)
			Expect(string(consumptionContents)).To(Equal(expectedConsumptionContents))
			Expect(consumptionDataPath).To(Equal(expectedConsumptionDataPath))

			expectedMetadataPath := filepath.Join(data.ConsumptionCollectorDataSetId, data.MetadataFileName)
			metadataContents, metadataPath := tarWriter.AddFileArgsForCall(3)
			Expect(metadataPath).To(Equal(expectedMetadataPath))

			var metadata data.Metadata
			Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
			Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
			Expect(metadata.EnvType).To(Equal(envType))
			Expect(metadata.FileDigests).To(ConsistOf(
				data.FileDigest{Name: consumptionData.Name(), MimeType: consumptionData.MimeType(), MD5Checksum: chContentMd5, ProductType: consumptionData.Type(), DataType: consumptionData.DataType()},
			))

			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("returns an error when the consumption collection errors", func() {
			consumptionDataCollector.CollectReturns(consumption.Data{}, errors.New("collecting is hard"))

			err := collectorWithConsumption.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(UsageCollectFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
		})

		It("returns an error when reading the consumption data content fails", func() {
			failingReader := new(operationsfakes.FakeReader)
			failingReader.ReadReturns(0, errors.New("reading is hard"))
			failingData := consumption.NewData(failingReader, "app-instances")
			consumptionDataCollector.CollectReturns(failingData, nil)

			err := collectorWithConsumption.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("reading is hard")))
		})

		It("returns an error when adding consumption data to the tar file fails", func() {
			usageData := consumption.NewData(strings.NewReader(""), "app-instances")
			consumptionDataCollector.CollectReturns(usageData, nil)
			tarWriter.AddFileStub = func(content []byte, filePath string) error {
				if filePath == filepath.Join(data.ConsumptionCollectorDataSetId, usageData.Name()) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithConsumption.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})

		It("returns an error when adding the metadata to the tar file fails", func() {
			consumptionDataCollector.CollectReturns(consumption.NewData(strings.NewReader(""), "app-instance"), nil)
			tarWriter.AddFileStub = func(contents []byte, filePath string) error {
				if filePath == filepath.Join(data.ConsumptionCollectorDataSetId, data.MetadataFileName) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithConsumption.Collect("", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})

	})

})

//go:generate counterfeiter . reader
type reader interface {
	io.Reader
}
