package file

import (
	"archive/tar"
	"crypto/md5"
	"encoding/base64"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

const (
	UnableToFindFileFormat  = "Could not find file %s in tar archive"
	UnexpectedFormatMessage = "Expected tar format"
	UnableToReadFileFormat  = "Unable to read file %s from tar file %s"
)

type TarReader struct {
	sourceTar io.ReadSeeker
}

func NewTarReader(sourceTar io.ReadSeeker) *TarReader {
	return &TarReader{sourceTar: sourceTar}
}

func (tr *TarReader) ReadFile(fileName string) ([]byte, error) {
	tr.sourceTar.Seek(0, 0)
	reader := tar.NewReader(tr.sourceTar)

	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			return []byte{}, errors.Wrapf(err, UnableToFindFileFormat, fileName)
		}
		if err != nil {
			return []byte{}, errors.Wrap(err, UnexpectedFormatMessage)
		}

		if hdr.Name == fileName {
			break
		}
	}

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, errors.Wrapf(err, UnableToReadFileFormat, fileName, tr.sourceTar)
	}

	return contents, nil
}

func (tr *TarReader) FileMd5s() (map[string]string, error) {
	tr.sourceTar.Seek(0, 0)
	reader := tar.NewReader(tr.sourceTar)

	fileMd5s := map[string]string{}
	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return map[string]string{}, errors.Wrap(err, UnexpectedFormatMessage)
		}

		contents, err := ioutil.ReadAll(reader)
		if err != nil {
			return map[string]string{}, errors.Wrapf(err, UnableToReadFileFormat, hdr.Name, tr.sourceTar)
		}

		fileSum := md5.Sum(contents)
		fileMd5s[hdr.Name] = base64.StdEncoding.EncodeToString(fileSum[:])
	}

	return fileMd5s, nil
}
