package operations

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"path"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pivotal-cf/aqueduct-courier/consumption"

	"github.com/pivotal-cf/aqueduct-courier/credhub"

	"github.com/pivotal-cf/aqueduct-courier/coreconsumption"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"
	"github.com/pkg/errors"
)

const (
	OpsManagerCollectFailureMessage = "Failed collecting from Operations Manager"
	CredhubCollectFailureMessage    = "Failed collecting from Credhub"
	UsageCollectFailureMessage      = "Failed collecting from Usage Service"
	DataWriteFailureMessage         = "Failed writing data"
	ContentReadingFailureMessage    = "Failed to read content"
	UUIDGenerationErrorMessage      = "unable to generate UUID"
	CoreCountsCollectFailureMessage = "Failed collecting from Core Counting API"
)

//go:generate counterfeiter . omDataCollector
type omDataCollector interface {
	Collect() ([]opsmanager.Data, string, error)
}

//go:generate counterfeiter . credhubDataCollector
type credhubDataCollector interface {
	Collect() (credhub.Data, error)
}

//go:generate counterfeiter . consumptionDataCollector
type consumptionDataCollector interface {
	Collect() ([]consumption.Data, error)
}

//go:generate counterfeiter . coreConsumptionDataCollector
type coreConsumptionDataCollector interface {
	Collect() ([]coreconsumption.Data, error)
}

//go:generate counterfeiter . tarWriter
type tarWriter interface {
	AddFile([]byte, string) error
	Close() error
}

//go:generate counterfeiter . uuidProvider
type uuidProvider interface {
	NewV4() (uuid.UUID, error)
}

type collectedData interface {
	Name() string
	MimeType() string
	DataType() string
	Type() string
	Content() io.Reader
}

type CollectExecutor struct {
	opsmanagerDC        omDataCollector
	credhubDC           credhubDataCollector
	consumptionDC       consumptionDataCollector
	coreConsumptionDC   coreConsumptionDataCollector
	tarWriter           tarWriter
	uuidProvider        uuidProvider
	operationalDataOnly bool
}

func NewCollector(opsmanagerDC omDataCollector, credhubDC credhubDataCollector, consumptionDC consumptionDataCollector, coreConsumptionDC coreConsumptionDataCollector, tarWriter tarWriter, uuidProvider uuidProvider, operationalDataOnly bool) *CollectExecutor {
	return &CollectExecutor{opsmanagerDC: opsmanagerDC, credhubDC: credhubDC, consumptionDC: consumptionDC, coreConsumptionDC: coreConsumptionDC, tarWriter: tarWriter, uuidProvider: uuidProvider, operationalDataOnly: operationalDataOnly}
}

func (ce *CollectExecutor) Collect(envType, collectorVersion, foundationNickname string) error {
	defer ce.tarWriter.Close()

	collectionID, err := ce.uuidProvider.NewV4()
	if err != nil {
		return errors.Wrap(err, UUIDGenerationErrorMessage)
	}
	collectionIDAsString := collectionID.String()

	omDatas, foundationId, err := ce.opsmanagerDC.Collect()
	if err != nil {
		return errors.Wrap(err, OpsManagerCollectFailureMessage)
	}

	collectedAtTime := time.Now().UTC().Format(time.RFC3339)
	opsManagerMetadata := collector_tar.Metadata{
		CollectorVersion:   collectorVersion,
		EnvType:            envType,
		CollectionId:       collectionIDAsString,
		FoundationId:       foundationId,
		FoundationNickname: foundationNickname,
		CollectedAt:        collectedAtTime,
	}

	usageMetadata := collector_tar.Metadata{
		CollectorVersion:   collectorVersion,
		EnvType:            envType,
		CollectionId:       collectionIDAsString,
		FoundationId:       foundationId,
		FoundationNickname: foundationNickname,
		CollectedAt:        collectedAtTime,
	}

	coreCountsMetadata := collector_tar.Metadata{
		CollectorVersion:   collectorVersion,
		EnvType:            envType,
		CollectionId:       collectionIDAsString,
		FoundationId:       foundationId,
		FoundationNickname: foundationNickname,
		CollectedAt:        collectedAtTime,
	}

	for _, omData := range omDatas {
		err = ce.addData(omData, &opsManagerMetadata, collector_tar.OpsManagerCollectorDataSetId)
		if err != nil {
			return err
		}
	}

	if ce.credhubDC != nil {
		chData, err := ce.credhubDC.Collect()
		if err != nil {
			return errors.Wrap(err, CredhubCollectFailureMessage)
		}

		err = ce.addData(chData, &opsManagerMetadata, collector_tar.OpsManagerCollectorDataSetId)
		if err != nil {
			return err
		}
	}

	metadataContents, err := json.Marshal(opsManagerMetadata)
	if err != nil {
		return err
	}

	if !ce.operationalDataOnly {
		err = ce.tarWriter.AddFile(metadataContents, path.Join(collector_tar.OpsManagerCollectorDataSetId, collector_tar.MetadataFileName))
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}
	}

	if ce.consumptionDC != nil {
		usageData, err := ce.consumptionDC.Collect()
		if err != nil {
			return errors.Wrap(err, UsageCollectFailureMessage)
		}

		for _, consumptionData := range usageData {
			err = ce.addData(consumptionData, &usageMetadata, collector_tar.UsageServiceCollectorDataSetId)
			if err != nil {
				return err
			}
		}
		usageMetadataContents, err := json.Marshal(usageMetadata)
		if err != nil {
			return err
		}

		err = ce.tarWriter.AddFile(usageMetadataContents, path.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName))
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}
	}

	if ce.coreConsumptionDC != nil {
		coreCountsData, err := ce.coreConsumptionDC.Collect()
		if err != nil {
			// Do not fail when we are unable to collect Core Consumption data
			// This API is only supported in Ops Manager 2.10.58+ and 3.0.10+
			log.Println(errors.Wrap(err, CoreCountsCollectFailureMessage))
		} else {
			for _, coreConsumptionData := range coreCountsData {
				err = ce.addData(coreConsumptionData, &coreCountsMetadata, collector_tar.CoreConsumptionCollectorDataSetId)
				if err != nil {
					return err
				}
			}

			coreCountsMetadataContents, err := json.Marshal(coreCountsMetadata)
			if err != nil {
				return err
			}

			err = ce.tarWriter.AddFile(coreCountsMetadataContents, path.Join(collector_tar.CoreConsumptionCollectorDataSetId, collector_tar.MetadataFileName))
			if err != nil {
				return errors.Wrap(err, DataWriteFailureMessage)
			}
		}
	}

	return nil
}

func (ce *CollectExecutor) addData(collectedData collectedData, metadata *collector_tar.Metadata, dataSetType string) error {
	dataContents, err := io.ReadAll(collectedData.Content())
	if err != nil {
		return errors.Wrap(err, ContentReadingFailureMessage)
	}

	err = ce.tarWriter.AddFile(dataContents, path.Join(dataSetType, collectedData.Name()))
	if err != nil {
		return errors.Wrap(err, DataWriteFailureMessage)
	}

	md5Sum := md5.Sum([]byte(dataContents))
	metadata.FileDigests = append(metadata.FileDigests, collector_tar.FileDigest{
		Name:        collectedData.Name(),
		MimeType:    collectedData.MimeType(),
		ProductType: collectedData.Type(),
		DataType:    collectedData.DataType(),
		MD5Checksum: base64.StdEncoding.EncodeToString(md5Sum[:]),
	})
	return nil
}
