package ops

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
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

type Metadata struct {
	EnvType      string
	CollectedAt  string
	CollectionId string
	FileDigests  []FileDigest
}
type FileDigest struct {
	Name        string
	MimeType    string
	MD5Checksum string
}

func NewCollector(c dataCollector, tw tarWriter) CollectExecutor {
	return CollectExecutor{c: c, tw: tw}
}

func (ce CollectExecutor) Collect(envType string) error {
	defer ce.tw.Close()

	omData, err := ce.c.Collect()
	if err != nil {
		return errors.Wrap(err, CollectFailureMessage)
	}

	metadata := Metadata{
		EnvType:      envType,
		CollectionId: uuid.NewV4().String(),
		CollectedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	for _, data := range omData {
		dataContents, err := ioutil.ReadAll(data.Content())
		if err != nil {
			return errors.Wrap(err, ContentReadingFailureMessage)
		}

		err = ce.tw.AddFile(dataContents, data.Name())
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}

		metadata.FileDigests = append(metadata.FileDigests, FileDigest{
			Name:        data.Name(),
			MimeType:    data.MimeType(),
			MD5Checksum: base64.StdEncoding.EncodeToString(dataContents),
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
