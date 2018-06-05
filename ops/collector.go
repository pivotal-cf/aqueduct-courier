package ops

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

const (
	CollectFailureMessage        = "Failed collecting from Operations Manager"
	DataWriteFailureMessage      = "Failed writing data"
	ContentReadingFailureMessage = "Failed to read content"
	MetadataFileName             = "aqueduct_metadata"
)

//go:generate counterfeiter . dataCollector
type dataCollector interface {
	Collect() ([]opsmanager.Data, error)
}

//go:generate counterfeiter . tarWriter
type tarWriter interface {
	AddFile([]byte, string) error
	Close() error
}

type CollectExecutor struct {
	c  dataCollector
	tw tarWriter
}

func NewCollector(c dataCollector, tw tarWriter) CollectExecutor {
	return CollectExecutor{c: c, tw: tw}
}

func (ce CollectExecutor) Collect(envType string) error {
	defer ce.tw.Close()

	omDatas, err := ce.c.Collect()
	if err != nil {
		return errors.Wrap(err, CollectFailureMessage)
	}

	metadata := data.Metadata{
		EnvType:      envType,
		CollectionId: uuid.NewV4().String(),
		CollectedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	for _, omData := range omDatas {
		dataContents, err := ioutil.ReadAll(omData.Content())
		if err != nil {
			return errors.Wrap(err, ContentReadingFailureMessage)
		}

		err = ce.tw.AddFile(dataContents, omData.Name())
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}

		md5Sum := md5.Sum([]byte(dataContents))
		metadata.FileDigests = append(metadata.FileDigests, data.FileDigest{
			Name:        omData.Name(),
			MimeType:    omData.MimeType(),
			ProductType: omData.Type(),
			DataType:    omData.DataType(),
			MD5Checksum: base64.StdEncoding.EncodeToString(md5Sum[:]),
		})
	}
	metadataContents, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	err = ce.tw.AddFile(metadataContents, MetadataFileName)
	if err != nil {
		return errors.Wrap(err, DataWriteFailureMessage)
	}

	return nil
}
