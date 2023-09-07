package operations_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/consumption"
	"github.com/pivotal-cf/aqueduct-courier/coreconsumption"

	"github.com/pivotal-cf/aqueduct-courier/credhub"

	"github.com/gofrs/uuid"
	"github.com/pivotal-cf/aqueduct-courier/operations"
	. "github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/aqueduct-courier/operations/operationsfakes"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
)

var _ = Describe("DataCollector", func() {
	var (
		omDataCollector              *operationsfakes.FakeOmDataCollector
		tarWriter                    *operationsfakes.FakeTarWriter
		uuidProvider                 *operationsfakes.FakeUuidProvider
		uuidString                   = "cf736154-6fd5-47f4-8ca9-1b4a6fe451ad"
		collector                    *CollectExecutor
		collectorOperationalDataOnly *CollectExecutor
	)

	BeforeEach(func() {
		omDataCollector = new(operationsfakes.FakeOmDataCollector)
		tarWriter = new(operationsfakes.FakeTarWriter)
		uuidProvider = new(operationsfakes.FakeUuidProvider)
		uuidProvider.NewV4Stub = func() (uuid.UUID, error) {
			return uuid.FromString(uuidString)
		}

		collector = NewCollector(omDataCollector, nil, nil, nil, tarWriter, uuidProvider, false)
		collectorOperationalDataOnly = NewCollector(omDataCollector, nil, nil, nil, tarWriter, uuidProvider, true)
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
		foundationId := "p-bosh-guid-of-some-sort"
		foundationNickname := "some-nickname"
		omDataCollector.CollectReturns(dataToWrite, foundationId, nil)

		collectorVersion := "0.0.1-version"
		envType := "most-production"

		err := collector.Collect(envType, collectorVersion, foundationNickname)
		Expect(err).NotTo(HaveOccurred())

		Expect(tarWriter.AddFileCallCount()).To(Equal(3))

		expectedD1Path := path.Join(collector_tar.OpsManagerCollectorDataSetId, d1.Name())
		d1Contents, d1Path := tarWriter.AddFileArgsForCall(0)
		expectedD2Path := path.Join(collector_tar.OpsManagerCollectorDataSetId, d2.Name())
		d2Contents, d2Path := tarWriter.AddFileArgsForCall(1)

		Expect(string(d1Contents)).To(Equal(expectedD1Contents))
		Expect(d1Path).To(Equal(expectedD1Path))
		Expect(string(d2Contents)).To(Equal(expectedD2Contents))
		Expect(d2Path).To(Equal(expectedD2Path))

		expectedMetadataPath := path.Join(collector_tar.OpsManagerCollectorDataSetId, collector_tar.MetadataFileName)
		metadataContents, metadataPath := tarWriter.AddFileArgsForCall(2)
		Expect(metadataPath).To(Equal(expectedMetadataPath))

		var metadata collector_tar.Metadata
		Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
		Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
		Expect(metadata.EnvType).To(Equal(envType))
		Expect(metadata.FileDigests).To(ConsistOf(
			collector_tar.FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type(), DataType: d1.DataType()},
			collector_tar.FileDigest{Name: d2.Name(), MimeType: d2.MimeType(), MD5Checksum: d2ContentMd5, ProductType: d2.Type(), DataType: d2.DataType()},
		))
		Expect(metadata.FoundationId).To(Equal(foundationId))
		Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
		Expect(metadata.CollectionId).To(Equal(uuidString))
		collectedAtTime, err := time.Parse(time.RFC3339, metadata.CollectedAt)
		Expect(err).NotTo(HaveOccurred())
		Expect(collectedAtTime.Location()).To(Equal(time.UTC))
		Expect(collectedAtTime).To(BeTemporally("~", time.Now(), time.Minute))

		Expect(tarWriter.CloseCallCount()).To(Equal(1))
	})

	It("does not collect opsmanager data when --operational-data-only flag is passed", func() {
		expectedD1Contents := "d1-content"
		expectedD2Contents := "d2-content"
		d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
		d2 := opsmanager.NewData(strings.NewReader(expectedD2Contents), "d2", "better-kind")
		dataToWrite := []opsmanager.Data{d1, d2}
		foundationId := "p-bosh-guid-of-some-sort"
		foundationNickname := "some-nickname"
		omDataCollector.CollectReturns(dataToWrite, foundationId, nil)

		collectorVersion := "0.0.1-version"
		envType := "most-production"

		err := collectorOperationalDataOnly.Collect(envType, collectorVersion, foundationNickname)
		Expect(err).NotTo(HaveOccurred())

		Expect(tarWriter.AddFileCallCount()).To(Equal(2))

		expectedD1Path := path.Join(collector_tar.OpsManagerCollectorDataSetId, d1.Name())
		d1Contents, d1Path := tarWriter.AddFileArgsForCall(0)
		expectedD2Path := path.Join(collector_tar.OpsManagerCollectorDataSetId, d2.Name())
		d2Contents, d2Path := tarWriter.AddFileArgsForCall(1)

		Expect(string(d1Contents)).To(Equal(expectedD1Contents))
		Expect(d1Path).To(Equal(expectedD1Path))
		Expect(string(d2Contents)).To(Equal(expectedD2Contents))
		Expect(d2Path).To(Equal(expectedD2Path))
	})

	It("returns an error when the ops manager collection errors", func() {
		omDataCollector.CollectReturns([]opsmanager.Data{}, "", errors.New("collecting is hard"))

		err := collector.Collect("", "", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(OpsManagerCollectFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
	})

	It("returns an error when reading the ops manager data content fails", func() {
		failingReader := new(operationsfakes.FakeReader)
		failingReader.ReadReturns(0, errors.New("reading is hard"))
		failingData := opsmanager.NewData(failingReader, "d1", "best-kind")
		omDataCollector.CollectReturns([]opsmanager.Data{failingData}, "", nil)

		err := collector.Collect("", "", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("reading is hard")))
	})

	It("returns an error when adding ops manager data to the tar file fails", func() {
		data := opsmanager.NewData(strings.NewReader(""), "d1", "best-kind")
		omDataCollector.CollectReturns([]opsmanager.Data{data}, "", nil)
		tarWriter.AddFileReturnsOnCall(0, errors.New("tarring is hard"))

		err := collector.Collect("", "", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when adding the metadata to the tar file fails", func() {
		tarWriter.AddFileStub = func(contents []byte, filePath string) error {
			if filePath == path.Join(collector_tar.OpsManagerCollectorDataSetId, collector_tar.MetadataFileName) {
				return errors.New("tarring is hard")
			}
			return nil
		}
		err := collector.Collect("", "", "")
		Expect(tarWriter.CloseCallCount()).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
		Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
	})

	It("returns an error when a UUID cannot be generated", func() {
		uuidProvider.NewV4Returns(uuid.UUID{}, errors.New("generating a UUID is hard"))

		err := collector.Collect("", "", "")
		Expect(err).To(MatchError(ContainSubstring(operations.UUIDGenerationErrorMessage)))
		Expect(err).To(MatchError(ContainSubstring("generating a UUID is hard")))
	})

	Describe("credhub collection", func() {
		var (
			collectorWithCredhub                    *CollectExecutor
			collectorWithCredhubOperationalDataOnly *CollectExecutor
			credhubDataCollector                    *operationsfakes.FakeCredhubDataCollector
		)

		BeforeEach(func() {
			credhubDataCollector = new(operationsfakes.FakeCredhubDataCollector)
			collectorWithCredhub = NewCollector(omDataCollector, credhubDataCollector, nil, nil, tarWriter, uuidProvider, false)
			collectorWithCredhubOperationalDataOnly = NewCollector(omDataCollector, credhubDataCollector, nil, nil, tarWriter, uuidProvider, true)
		})

		It("collects credhub data and writes it", func() {
			expectedD1Contents := "d1-content"
			md5SumD1 := md5.Sum([]byte(expectedD1Contents))
			d1ContentMd5 := base64.StdEncoding.EncodeToString(md5SumD1[:])
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			omDataCollector.CollectReturns(dataToWrite, "", nil)

			expectedCHContents := "ch-content"
			md5SumCH := md5.Sum([]byte(expectedCHContents))
			chContentMd5 := base64.StdEncoding.EncodeToString(md5SumCH[:])
			chData := credhub.NewData(strings.NewReader(expectedCHContents))
			credhubDataCollector.CollectReturns(chData, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"
			foundationNickname := "some-nickname"

			err := collectorWithCredhub.Collect(envType, collectorVersion, foundationNickname)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(3))

			chContents, credhubDataPath := tarWriter.AddFileArgsForCall(1)
			Expect(string(chContents)).To(Equal(expectedCHContents))

			expectedCredhubDataPath := path.Join(collector_tar.OpsManagerCollectorDataSetId, chData.Name())
			Expect(credhubDataPath).To(Equal(expectedCredhubDataPath))

			expectedMetadataPath := path.Join(collector_tar.OpsManagerCollectorDataSetId, collector_tar.MetadataFileName)
			metadataContents, metadataPath := tarWriter.AddFileArgsForCall(2)

			Expect(metadataPath).To(Equal(expectedMetadataPath))
			var metadata collector_tar.Metadata
			Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
			Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
			Expect(metadata.EnvType).To(Equal(envType))
			Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
			Expect(metadata.FileDigests).To(ConsistOf(
				collector_tar.FileDigest{Name: d1.Name(), MimeType: d1.MimeType(), MD5Checksum: d1ContentMd5, ProductType: d1.Type(), DataType: d1.DataType()},
				collector_tar.FileDigest{Name: chData.Name(), MimeType: chData.MimeType(), MD5Checksum: chContentMd5, ProductType: chData.Type(), DataType: chData.DataType()},
			))

			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("does not collect opsmanager data when --operational-data-only flag is passed", func() {
			expectedD1Contents := "d1-content"
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			omDataCollector.CollectReturns(dataToWrite, "", nil)

			expectedCHContents := "ch-content"
			chData := credhub.NewData(strings.NewReader(expectedCHContents))
			credhubDataCollector.CollectReturns(chData, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"
			foundationNickname := "some-nickname"

			err := collectorWithCredhubOperationalDataOnly.Collect(envType, collectorVersion, foundationNickname)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(2))

			chContents, credhubDataPath := tarWriter.AddFileArgsForCall(1)
			Expect(string(chContents)).To(Equal(expectedCHContents))

			expectedCredhubDataPath := path.Join(collector_tar.OpsManagerCollectorDataSetId, chData.Name())
			Expect(credhubDataPath).To(Equal(expectedCredhubDataPath))
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("returns an error when the credhub collection errors", func() {
			credhubDataCollector.CollectReturns(credhub.Data{}, errors.New("collecting is hard"))

			err := collectorWithCredhub.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(CredhubCollectFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
		})

		It("returns an error when reading the credhub data content fails", func() {
			failingReader := new(operationsfakes.FakeReader)
			failingReader.ReadReturns(0, errors.New("reading is hard"))
			failingData := credhub.NewData(failingReader)
			credhubDataCollector.CollectReturns(failingData, nil)

			err := collectorWithCredhub.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("reading is hard")))
		})

		It("returns an error when adding credhub data to the tar file fails", func() {
			credhubData := credhub.NewData(strings.NewReader(""))
			credhubDataCollector.CollectReturns(credhubData, nil)
			tarWriter.AddFileStub = func(content []byte, filePath string) error {
				if filePath == path.Join(collector_tar.OpsManagerCollectorDataSetId, credhubData.Name()) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithCredhub.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})
	})

	Describe("consumption collection", func() {
		var (
			collectorWithConsumption                    *CollectExecutor
			consumptionDataCollector                    *operationsfakes.FakeConsumptionDataCollector
			collectorWithConsumptionOperationalDataOnly *CollectExecutor
		)

		BeforeEach(func() {
			consumptionDataCollector = new(operationsfakes.FakeConsumptionDataCollector)
			collectorWithConsumption = NewCollector(omDataCollector, nil, consumptionDataCollector, nil, tarWriter, uuidProvider, false)
			collectorWithConsumptionOperationalDataOnly = NewCollector(omDataCollector, nil, consumptionDataCollector, nil, tarWriter, uuidProvider, true)
		})

		It("collects consumption data and writes it", func() {
			expectedD1Contents := "d1-content"
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			foundationId := "p-bosh-guid-of-some-sort"
			omDataCollector.CollectReturns(dataToWrite, foundationId, nil)

			expectedAppUsageConsumptionContents := "consumption-app-usage-content"
			md5sum := md5.Sum([]byte(expectedAppUsageConsumptionContents))
			appUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
			appUsageConsumptionData := consumption.NewData(strings.NewReader(expectedAppUsageConsumptionContents), "app-instances")

			expectedServiceUsageConsumptionContents := "consumption-service-usage-content"
			md5sum = md5.Sum([]byte(expectedServiceUsageConsumptionContents))
			serviceUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
			serviceUsageConsumptionData := consumption.NewData(strings.NewReader(expectedServiceUsageConsumptionContents), "service-instances")

			consumptionDataCollector.CollectReturns([]consumption.Data{appUsageConsumptionData, serviceUsageConsumptionData}, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"
			foundationNickname := "some-nickname"

			err := collectorWithConsumption.Collect(envType, collectorVersion, foundationNickname)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(5))

			expectedAppUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, appUsageConsumptionData.Name())
			appUsageConsumptionContents, appUsageConsumptionDataPath := tarWriter.AddFileArgsForCall(2)
			Expect(string(appUsageConsumptionContents)).To(Equal(expectedAppUsageConsumptionContents))
			Expect(appUsageConsumptionDataPath).To(Equal(expectedAppUsageConsumptionDataPath))

			expectedServiceUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, serviceUsageConsumptionData.Name())
			serviceUsageConsumptionContents, serviceConsumptionDataPath := tarWriter.AddFileArgsForCall(3)
			Expect(string(serviceUsageConsumptionContents)).To(Equal(expectedServiceUsageConsumptionContents))
			Expect(serviceConsumptionDataPath).To(Equal(expectedServiceUsageConsumptionDataPath))

			expectedMetadataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName)
			metadataContents, metadataPath := tarWriter.AddFileArgsForCall(4)
			Expect(metadataPath).To(Equal(expectedMetadataPath))

			var metadata collector_tar.Metadata
			Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
			Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
			Expect(metadata.CollectionId).To(Equal(uuidString))
			Expect(metadata.FoundationId).To(Equal(foundationId))
			Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
			Expect(metadata.EnvType).To(Equal(envType))
			Expect(metadata.FileDigests).To(ConsistOf(
				collector_tar.FileDigest{Name: appUsageConsumptionData.Name(), MimeType: appUsageConsumptionData.MimeType(), MD5Checksum: appUsageContentMd5, ProductType: appUsageConsumptionData.Type(), DataType: appUsageConsumptionData.DataType()},
				collector_tar.FileDigest{Name: serviceUsageConsumptionData.Name(), MimeType: serviceUsageConsumptionData.MimeType(), MD5Checksum: serviceUsageContentMd5, ProductType: serviceUsageConsumptionData.Type(), DataType: serviceUsageConsumptionData.DataType()},
			))

			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("does not collect opsmanager data when --operational-data-only flag is passed", func() {
			expectedD1Contents := "d1-content"
			d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
			dataToWrite := []opsmanager.Data{d1}
			foundationId := "p-bosh-guid-of-some-sort"
			omDataCollector.CollectReturns(dataToWrite, foundationId, nil)

			expectedAppUsageConsumptionContents := "consumption-app-usage-content"
			md5sum := md5.Sum([]byte(expectedAppUsageConsumptionContents))
			appUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
			appUsageConsumptionData := consumption.NewData(strings.NewReader(expectedAppUsageConsumptionContents), "app-instances")

			expectedServiceUsageConsumptionContents := "consumption-service-usage-content"
			md5sum = md5.Sum([]byte(expectedServiceUsageConsumptionContents))
			serviceUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
			serviceUsageConsumptionData := consumption.NewData(strings.NewReader(expectedServiceUsageConsumptionContents), "service-instances")

			consumptionDataCollector.CollectReturns([]consumption.Data{appUsageConsumptionData, serviceUsageConsumptionData}, nil)

			collectorVersion := "0.0.1-version"
			envType := "most-production"
			foundationNickname := "some-nickname"

			err := collectorWithConsumptionOperationalDataOnly.Collect(envType, collectorVersion, foundationNickname)
			Expect(err).NotTo(HaveOccurred())

			Expect(tarWriter.AddFileCallCount()).To(Equal(4))

			expectedAppUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, appUsageConsumptionData.Name())
			appUsageConsumptionContents, appUsageConsumptionDataPath := tarWriter.AddFileArgsForCall(1)
			Expect(string(appUsageConsumptionContents)).To(Equal(expectedAppUsageConsumptionContents))
			Expect(appUsageConsumptionDataPath).To(Equal(expectedAppUsageConsumptionDataPath))

			expectedServiceUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, serviceUsageConsumptionData.Name())
			serviceUsageConsumptionContents, serviceConsumptionDataPath := tarWriter.AddFileArgsForCall(2)
			Expect(string(serviceUsageConsumptionContents)).To(Equal(expectedServiceUsageConsumptionContents))
			Expect(serviceConsumptionDataPath).To(Equal(expectedServiceUsageConsumptionDataPath))

			expectedMetadataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName)
			metadataContents, metadataPath := tarWriter.AddFileArgsForCall(3)
			Expect(metadataPath).To(Equal(expectedMetadataPath))

			var metadata collector_tar.Metadata
			Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
			Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
			Expect(metadata.CollectionId).To(Equal(uuidString))
			Expect(metadata.FoundationId).To(Equal(foundationId))
			Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
			Expect(metadata.EnvType).To(Equal(envType))
			Expect(metadata.FileDigests).To(ConsistOf(
				collector_tar.FileDigest{Name: appUsageConsumptionData.Name(), MimeType: appUsageConsumptionData.MimeType(), MD5Checksum: appUsageContentMd5, ProductType: appUsageConsumptionData.Type(), DataType: appUsageConsumptionData.DataType()},
				collector_tar.FileDigest{Name: serviceUsageConsumptionData.Name(), MimeType: serviceUsageConsumptionData.MimeType(), MD5Checksum: serviceUsageContentMd5, ProductType: serviceUsageConsumptionData.Type(), DataType: serviceUsageConsumptionData.DataType()},
			))

			Expect(tarWriter.CloseCallCount()).To(Equal(1))
		})

		It("returns an error when the consumption collection errors", func() {
			consumptionDataCollector.CollectReturns([]consumption.Data{}, errors.New("collecting is hard"))

			err := collectorWithConsumption.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(UsageCollectFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("collecting is hard")))
		})

		It("returns an error when reading the consumption data content fails", func() {
			failingReader := new(operationsfakes.FakeReader)
			failingReader.ReadReturns(0, errors.New("reading is hard"))
			failingData := consumption.NewData(failingReader, "app-instances")
			consumptionDataCollector.CollectReturns([]consumption.Data{failingData}, nil)

			err := collectorWithConsumption.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(ContentReadingFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("reading is hard")))
		})

		It("returns an error when adding consumption data to the tar file fails", func() {
			usageData := consumption.NewData(strings.NewReader(""), "app-instances")
			consumptionDataCollector.CollectReturns([]consumption.Data{usageData}, nil)
			tarWriter.AddFileStub = func(content []byte, filePath string) error {
				if filePath == path.Join(collector_tar.UsageServiceCollectorDataSetId, usageData.Name()) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithConsumption.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})

		It("returns an error when adding the metadata to the tar file fails", func() {
			consumptionDataCollector.CollectReturns([]consumption.Data{consumption.NewData(strings.NewReader(""), "app-instance")}, nil)
			tarWriter.AddFileStub = func(contents []byte, filePath string) error {
				if filePath == path.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName) {
					return errors.New("tarring is hard")
				}
				return nil
			}

			err := collectorWithConsumption.Collect("", "", "")
			Expect(tarWriter.CloseCallCount()).To(Equal(1))
			Expect(err).To(MatchError(ContainSubstring(DataWriteFailureMessage)))
			Expect(err).To(MatchError(ContainSubstring("tarring is hard")))
		})

		Describe("--operational-data-only flag", func() {
			It("does not collect opsmanager data", func() {
				expectedD1Contents := "d1-content"
				d1 := opsmanager.NewData(strings.NewReader(expectedD1Contents), "d1", "best-kind")
				dataToWrite := []opsmanager.Data{d1}
				foundationId := "p-bosh-guid-of-some-sort"
				omDataCollector.CollectReturns(dataToWrite, foundationId, nil)

				expectedAppUsageConsumptionContents := "consumption-app-usage-content"
				md5sum := md5.Sum([]byte(expectedAppUsageConsumptionContents))
				appUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
				appUsageConsumptionData := consumption.NewData(strings.NewReader(expectedAppUsageConsumptionContents), "app-instances")

				expectedServiceUsageConsumptionContents := "consumption-service-usage-content"
				md5sum = md5.Sum([]byte(expectedServiceUsageConsumptionContents))
				serviceUsageContentMd5 := base64.StdEncoding.EncodeToString(md5sum[:])
				serviceUsageConsumptionData := consumption.NewData(strings.NewReader(expectedServiceUsageConsumptionContents), "service-instances")

				consumptionDataCollector.CollectReturns([]consumption.Data{appUsageConsumptionData, serviceUsageConsumptionData}, nil)

				collectorVersion := "0.0.1-version"
				envType := "most-production"
				foundationNickname := "some-nickname"

				err := collectorWithConsumption.Collect(envType, collectorVersion, foundationNickname)
				Expect(err).NotTo(HaveOccurred())

				Expect(tarWriter.AddFileCallCount()).To(Equal(5))

				expectedAppUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, appUsageConsumptionData.Name())
				appUsageConsumptionContents, appUsageConsumptionDataPath := tarWriter.AddFileArgsForCall(2)
				Expect(string(appUsageConsumptionContents)).To(Equal(expectedAppUsageConsumptionContents))
				Expect(appUsageConsumptionDataPath).To(Equal(expectedAppUsageConsumptionDataPath))

				expectedServiceUsageConsumptionDataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, serviceUsageConsumptionData.Name())
				serviceUsageConsumptionContents, serviceConsumptionDataPath := tarWriter.AddFileArgsForCall(3)
				Expect(string(serviceUsageConsumptionContents)).To(Equal(expectedServiceUsageConsumptionContents))
				Expect(serviceConsumptionDataPath).To(Equal(expectedServiceUsageConsumptionDataPath))

				expectedMetadataPath := path.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName)
				metadataContents, metadataPath := tarWriter.AddFileArgsForCall(4)
				Expect(metadataPath).To(Equal(expectedMetadataPath))

				var metadata collector_tar.Metadata
				Expect(json.Unmarshal(metadataContents, &metadata)).To(Succeed())
				Expect(metadata.CollectorVersion).To(Equal(collectorVersion))
				Expect(metadata.CollectionId).To(Equal(uuidString))
				Expect(metadata.FoundationId).To(Equal(foundationId))
				Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
				Expect(metadata.EnvType).To(Equal(envType))
				Expect(metadata.FileDigests).To(ConsistOf(
					collector_tar.FileDigest{Name: appUsageConsumptionData.Name(), MimeType: appUsageConsumptionData.MimeType(), MD5Checksum: appUsageContentMd5, ProductType: appUsageConsumptionData.Type(), DataType: appUsageConsumptionData.DataType()},
					collector_tar.FileDigest{Name: serviceUsageConsumptionData.Name(), MimeType: serviceUsageConsumptionData.MimeType(), MD5Checksum: serviceUsageContentMd5, ProductType: serviceUsageConsumptionData.Type(), DataType: serviceUsageConsumptionData.DataType()},
				))

				Expect(tarWriter.CloseCallCount()).To(Equal(1))
			})
		})
	})

	Describe("core consumption collection", func() {
		var (
			coreConsumptionDC  *operationsfakes.FakeCoreConsumptionDataCollector
			collectorOldOpsMan *CollectExecutor
		)

		BeforeEach(func() {
			coreConsumptionDC = new(operationsfakes.FakeCoreConsumptionDataCollector)
			coreConsumptionDC.CollectReturns([]coreconsumption.Data{}, errors.New("Can't collect Core Consumption"))
			collectorOldOpsMan = NewCollector(omDataCollector, nil, nil, coreConsumptionDC, tarWriter, uuidProvider, false)
		})

		It("Does not fail when collect fails", func() {
			err := collectorOldOpsMan.Collect("", "", "")
			Expect(err).To(BeNil())
		})
	})

})
