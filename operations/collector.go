package operations

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pivotal-cf/aqueduct-courier/consumption"

	"github.com/pivotal-cf/aqueduct-courier/credhub"

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
	opsmanagerDC  omDataCollector
	credhubDC     credhubDataCollector
	consumptionDC consumptionDataCollector
	tarWriter     tarWriter
	uuidProvider  uuidProvider
}

func NewCollector(opsmanagerDC omDataCollector, credhubDC credhubDataCollector, consumptionDC consumptionDataCollector, tarWriter tarWriter, uuidProvider uuidProvider) *CollectExecutor {
	return &CollectExecutor{opsmanagerDC: opsmanagerDC, credhubDC: credhubDC, consumptionDC: consumptionDC, tarWriter: tarWriter, uuidProvider: uuidProvider}
}

func (ce *CollectExecutor) Collect(envType, collectorVersion string) error {
	defer ce.tarWriter.Close()

	collectionID, err := ce.uuidProvider.NewV4()
	if err != nil {
		return errors.Wrap(err, UUIDGenerationErrorMessage)
	}

	omDatas, foundationId, err := ce.opsmanagerDC.Collect()
	if err != nil {
		return errors.Wrap(err, OpsManagerCollectFailureMessage)
	}

	opsManagerMetadata := collector_tar.Metadata{
		CollectorVersion: collectorVersion,
		EnvType:          envType,
		CollectionId:     collectionID.String(),
		FoundationId:     foundationId,
		CollectedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	usageMetadata := collector_tar.Metadata{
		CollectorVersion: collectorVersion,
		EnvType:          envType,
		CollectionId:     opsManagerMetadata.CollectionId,
		FoundationId:     foundationId,
		CollectedAt:      opsManagerMetadata.CollectedAt,
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
	err = ce.tarWriter.AddFile(metadataContents, filepath.Join(collector_tar.OpsManagerCollectorDataSetId, collector_tar.MetadataFileName))
	if err != nil {
		return errors.Wrap(err, DataWriteFailureMessage)
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

		err = ce.tarWriter.AddFile(usageMetadataContents, filepath.Join(collector_tar.UsageServiceCollectorDataSetId, collector_tar.MetadataFileName))
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}
	}

	return nil
}

func (ce *CollectExecutor) addData(collectedData collectedData, metadata *collector_tar.Metadata, dataSetType string) error {
	dataContents, err := ioutil.ReadAll(collectedData.Content())
	if err != nil {
		return errors.Wrap(err, ContentReadingFailureMessage)
	}

	err = ce.tarWriter.AddFile(dataContents, filepath.Join(dataSetType, collectedData.Name()))
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
