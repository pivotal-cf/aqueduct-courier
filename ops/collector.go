package ops

import (
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

const (
	CollectFailureMessage        = "Failed collecting from Operations Manager"
	DirCreateFailureMessage      = "Creating output directory failed"
	DataWriteFailureMessage      = "Failed writing data"
	UUIDGenerationFailureMessage = "Failed to generate unique collection ID"
)

//go:generate counterfeiter . dataCollector
type dataCollector interface {
	Collect() ([]opsmanager.Data, error)
}

//go:generate counterfeiter . writer
type writer interface {
	Write(file.Data, string, string) error
	Mkdir(string) (string, error)
}

type CollectExecutor struct {
	c dataCollector
	w writer
}

func NewCollector(c dataCollector, w writer) CollectExecutor {
	return CollectExecutor{c: c, w: w}
}

func (ce CollectExecutor) Collect(path string) error {
	omData, err := ce.c.Collect()
	if err != nil {
		return errors.Wrap(err, CollectFailureMessage)
	}

	collectionId, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, UUIDGenerationFailureMessage)
	}

	outputFolderPath, err := ce.w.Mkdir(path)
	if err != nil {
		return errors.Wrap(err, DirCreateFailureMessage)
	}
	for _, data := range omData {
		err = ce.w.Write(data, outputFolderPath, collectionId.String())
		if err != nil {
			return errors.Wrap(err, DataWriteFailureMessage)
		}
	}
	return nil
}
