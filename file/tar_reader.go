package file

import (
	"archive/tar"
	"crypto/md5"
	"encoding/base64"
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

func (tr *TarReader) TarFilePath() string {
	return tr.sourceTarFile
}

func (tr *TarReader) ReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(tr.sourceTarFile)
	if err != nil {
		return []byte{}, errors.Wrapf(err, OpenTarFileFailureFormat, tr.sourceTarFile)
	}
	defer file.Close()

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

func (tr *TarReader) FileMd5s() (map[string]string, error) {
	file, err := os.Open(tr.sourceTarFile)
	if err != nil {
		return map[string]string{}, errors.Wrapf(err, OpenTarFileFailureFormat, tr.sourceTarFile)
	}
	defer file.Close()

	reader := tar.NewReader(file)

	fileMd5s := map[string]string{}
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return map[string]string{}, errors.Wrapf(err, UnexpectedFileTypeFormat, tr.sourceTarFile)
		}

		contents, err := ioutil.ReadAll(reader)
		if err != nil {
			return map[string]string{}, errors.Wrapf(err, UnableToReadFileFormat, hdr.Name, tr.sourceTarFile)
		}

		fileSum := md5.Sum(contents)
		fileMd5s[hdr.Name] = base64.StdEncoding.EncodeToString(fileSum[:])
	}

	return fileMd5s, nil
}
