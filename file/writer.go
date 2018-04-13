package file

import (
	"fmt"
	"io"
	"path/filepath"

	"os"

	"github.com/pkg/errors"
)

const (
	CreateErrorFormat         = "Failed to create file for %s"
	ContentWritingErrorFormat = "Failed to write content for %s"
)

type Writer struct{}

//go:generate counterfeiter . data
type data interface {
	Name() string
	Content() io.Reader
	ContentType() string
}

func (w Writer) Write(d data, dir string) error {
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
