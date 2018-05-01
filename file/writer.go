package file

import (
	"fmt"
	"io"
	"path/filepath"

	"os"

	"time"

	"io/ioutil"

	"encoding/json"

	"crypto/md5"

	"github.com/pkg/errors"
)

const (
	OutputDirPrefix           = "FoundationDetails_"
	ContentWritingErrorFormat = "Failed to write file for %s"
	ContentReadingErrorFormat = "Failed to read content for %s"
	MetadataFileName          = "aqueduct_metadata.json"
)

type Writer struct {
	metadata Metadata
}

type Metadata struct {
	CollectedAt string
	Files       []FileData
}

type FileData struct {
	Name        string
	ContentType string
	Checksum    string
}

//go:generate counterfeiter . Data
type Data interface {
	Name() string
	Content() io.Reader
	ContentType() string
}

func (w *Writer) Write(d Data, dir string) error {
	dataContents, err := ioutil.ReadAll(d.Content())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentReadingErrorFormat, d.Name()))
	}

	err = ioutil.WriteFile(
		filepath.Join(dir, fmt.Sprintf("%s.%s", d.Name(), d.ContentType())),
		dataContents,
		0644,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentWritingErrorFormat, d.Name()))
	}

	err = w.writeOrUpdateMetadata(d, dataContents, dir)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) Mkdir(dirPrefix string) (string, error) {
	timeString := time.Now().UTC().Unix()
	outputFolderPath := filepath.Join(dirPrefix, fmt.Sprintf("%s%d", OutputDirPrefix, timeString))
	err := os.Mkdir(outputFolderPath, 0755)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed creating directory %s:", outputFolderPath))
	}

	return outputFolderPath, nil
}

func (w *Writer) writeOrUpdateMetadata(d Data, content []byte, dir string) error {
	w.metadata.CollectedAt = time.Now().UTC().Format(time.RFC3339)
	fileMetadata := FileData{
		Name:        d.Name(),
		ContentType: d.ContentType(),
		Checksum:    fmt.Sprintf("%x", md5.Sum(content)),
	}
	w.metadata.Files = append(w.metadata.Files, fileMetadata)

	metadataBytes, err := json.Marshal(w.metadata)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, MetadataFileName), metadataBytes, 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentWritingErrorFormat, MetadataFileName))
	}
	return nil
}
