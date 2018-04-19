package file

import (
	"fmt"
	"io"
	"path/filepath"

	"os"

	"time"

	"github.com/pkg/errors"
)

const (
	OutputDirPrefix           = "FoundationDetails_"
	CreateErrorFormat         = "Failed to create file for %s"
	ContentWritingErrorFormat = "Failed to write content for %s"
)

type Writer struct{}

//go:generate counterfeiter . Data
type Data interface {
	Name() string
	Content() io.Reader
	ContentType() string
}

func (w Writer) Write(d Data, dir string) error {
	file, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s.%s", d.Name(), d.ContentType())))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(CreateErrorFormat, d.Name()))
	}
	defer file.Close()
	_, err = io.Copy(file, d.Content())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(ContentWritingErrorFormat, d.Name()))
	}

	return nil
}

func (w Writer) Mkdir(dirPrefix string) (string, error) {
	timeString := time.Now().UTC().Format(time.RFC3339)
	outputFolderPath := filepath.Join(dirPrefix, fmt.Sprintf("%s%s", OutputDirPrefix, timeString))
	err := os.Mkdir(outputFolderPath, 0755)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed creating directory %s:", outputFolderPath))
	}
	return outputFolderPath, nil
}
