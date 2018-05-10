package file

import (
	"archive/tar"
	"os"

	"github.com/pkg/errors"
)

const (
	CreateTarFileFailureFormat    = "Could not create tar file %s"
	WriteTarHeaderFailureFormat   = "Could not write tar header for %s"
	WriteTarContentsFailureFormat = "Could not write tar header for %s"
	CloseWriterFailureMessage     = "Failed to close writer"
)

type TarWriter struct {
	writer  *tar.Writer
	tarFile *os.File
}

func NewTarWriter(fileToCreate string) (*TarWriter, error) {
	tarFile, err := os.Create(fileToCreate)
	if err != nil {
		return nil, errors.Wrapf(err, CreateTarFileFailureFormat, fileToCreate)
	}
	return &TarWriter{writer: tar.NewWriter(tarFile), tarFile: tarFile}, nil
}

func (tw *TarWriter) AddFile(contents []byte, fileName string) error {
	fileHeader := &tar.Header{
		Name: fileName,
		Size: int64(len(contents)),
		Mode: 0644,
	}
	if err := tw.writer.WriteHeader(fileHeader); err != nil {
		return errors.Wrapf(err, WriteTarHeaderFailureFormat, fileName)
	}

	if _, err := tw.writer.Write(contents); err != nil {
		return errors.Wrapf(err, WriteTarContentsFailureFormat, fileName)
	}

	return nil
}

func (tw *TarWriter) Close() error {
	if err := tw.writer.Close(); err != nil {
		return errors.Wrap(err, CloseWriterFailureMessage)
	}

	// close tar file as well for windows
	if err := tw.tarFile.Close(); err != nil {
		return errors.Wrap(err, CloseWriterFailureMessage)
	}

	return nil
}
