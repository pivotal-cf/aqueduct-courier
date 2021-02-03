package tar

import (
	"archive/tar"
	"io"

	"github.com/pkg/errors"
)

const (
	WriteTarHeaderFailureFormat   = "Could not write tar header for %s"
	WriteTarContentsFailureFormat = "Could not write tar header for %s"
	CloseWriterFailureMessage     = "Failed to close writer"
)

type TarWriter struct {
	writer *tar.Writer
}

func NewTarWriter(writer io.Writer) *TarWriter {
	return &TarWriter{writer: tar.NewWriter(writer)}
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

	return nil
}
