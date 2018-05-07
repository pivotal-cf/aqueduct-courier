package file

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

const (
	OutputDirPrefix           = "FoundationDetails_"
	ContentWritingErrorFormat = "Failed to write file for %s"
	ContentReadingErrorFormat = "Failed to read content for %s"
	MetadataFileName          = "aqueduct_metadata"
)

type Writer struct {
	metadata Metadata
}

type Metadata struct {
	EnvType     string
	CollectedAt string
	FileDigests []Digest
}

type Digest struct {
	Name        string
	MimeType    string
	MD5Checksum string
}

//go:generate counterfeiter . Data
type Data interface {
	Name() string
	Content() io.Reader
	MimeType() string
}

func NewWriter(envType string) *Writer {
	return &Writer{
		metadata: Metadata{
			EnvType: envType,
		},
	}
}

func (w *Writer) Write(d Data, dir string) error {
	dataContents, err := ioutil.ReadAll(d.Content())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentReadingErrorFormat, d.Name()))
	}

	err = ioutil.WriteFile(
		filepath.Join(dir, d.Name()),
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
	md5Sum := md5.Sum(content)
	fileMetadata := Digest{
		Name:        d.Name(),
		MimeType:    d.MimeType(),
		MD5Checksum: base64.StdEncoding.EncodeToString(md5Sum[:]),
	}
	w.metadata.FileDigests = append(w.metadata.FileDigests, fileMetadata)

	metadataBytes, err := json.Marshal(w.metadata)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(dir, MetadataFileName), metadataBytes, 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentWritingErrorFormat, MetadataFileName))
	}
	return nil
}
