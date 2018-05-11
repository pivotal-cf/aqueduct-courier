package file

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

const (
	OpenTarFileFailureFormat = "Could not open tar file %s"
	UnableToFindFileFormat   = "Could not find file %s in tar archive"
	UnexpectedFileTypeFormat = "Expected %s to be a tar file"
	UnableToReadFileFormat   = "Unable to read file %s from tar file %s"
)

type TarReader struct {
	sourceTarFile string
}

func NewTarReader(sourceTarfile string) *TarReader {
	return &TarReader{sourceTarFile: sourceTarfile}
}

func (tr *TarReader) ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(tr.sourceTarFile)
	if err != nil {
		return []byte{}, errors.Wrapf(err, OpenTarFileFailureFormat, tr.sourceTarFile)
	}
	reader := tar.NewReader(file)

	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			return []byte{}, errors.Wrapf(err, UnableToFindFileFormat, fileName)
		}
		if err != nil {
			return []byte{}, errors.Wrapf(err, UnexpectedFileTypeFormat, tr.sourceTarFile)
		}

		if hdr.Name == fileName {
			break
		}
	}

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, errors.Wrapf(err, UnableToReadFileFormat, fileName, tr.sourceTarFile)
	}

	return contents, nil
}
