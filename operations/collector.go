package operations

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/credhub"

	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const (
	OpsManagerCollectFailureMessage = "Failed collecting from Operations Manager"
	CredhubCollectFailureMessage    = "Failed collecting from Credhub"
	DataWriteFailureMessage         = "Failed writing data"
	ContentReadingFailureMessage    = "Failed to read content"
)

//go:generate counterfeiter . omDataCollector
type omDataCollector interface {
	Collect() ([]opsmanager.Data, error)
}

//go:generate counterfeiter . credhubDataCollector
type credhubDataCollector interface {
	Collect() (credhub.Data, error)
}

//go:generate counterfeiter . tarWriter
type tarWriter interface {
	AddFile([]byte, string) error
	Close() error
}

type CollectExecutor struct {
	omDC omDataCollector
	chDC credhubDataCollector
	tw   tarWriter
}

type collectedData interface {
	Name() string
	MimeType() string
	DataType() string
	Type() string
	Content() io.Reader
}

func NewCollector(omDC omDataCollector, chDC credhubDataCollector, tw tarWriter) CollectExecutor {
	return CollectExecutor{omDC: omDC, chDC: chDC, tw: tw}
}

func (ce CollectExecutor) Collect(envType, collectorVersion string) error {
	defer ce.tw.Close()

	omDatas, err := ce.omDC.Collect()
	if err != nil {
		return errors.Wrap(err, OpsManagerCollectFailureMessage)
	}

	metadata := data.Metadata{
		CollectorVersion: collectorVersion,
		EnvType:          envType,
		CollectionId:     uuid.NewV4().String(),
		CollectedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	for _, omData := range omDatas {
		err = ce.addData(omData, &metadata)
		if err != nil {
			return err
		}
	}

	if ce.chDC != nil {
		chData, err := ce.chDC.Collect()
		if err != nil {
			return errors.Wrap(err, CredhubCollectFailureMessage)
		}

		err = ce.addData(chData, &metadata)
		if err != nil {
			return err
		}
	}

	metadataContents, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	err = ce.tw.AddFile(metadataContents, data.MetadataFileName)
	if err != nil {
		return errors.Wrap(err, DataWriteFailureMessage)
	}

	return nil
}

func (ce CollectExecutor) addData(cData collectedData, metadata *data.Metadata) error {
	dataContents, err := ioutil.ReadAll(cData.Content())
	if err != nil {
		return errors.Wrap(err, ContentReadingFailureMessage)
	}

	err = ce.tw.AddFile(dataContents, cData.Name())
	if err != nil {
		return errors.Wrap(err, DataWriteFailureMessage)
	}

	md5Sum := md5.Sum([]byte(dataContents))
	metadata.FileDigests = append(metadata.FileDigests, data.FileDigest{
		Name:        cData.Name(),
		MimeType:    cData.MimeType(),
		ProductType: cData.Type(),
		DataType:    cData.DataType(),
		MD5Checksum: base64.StdEncoding.EncodeToString(md5Sum[:]),
	})
	return nil
}
