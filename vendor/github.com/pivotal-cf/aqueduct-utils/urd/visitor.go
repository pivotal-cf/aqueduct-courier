package urd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
)

const (
	ListFilesErrorMessage              = "Failed to list files"
	CleanupFileErrorFormat             = "Failed to clean up downloaded file: %s"
	CreateDataFileErrorFormat          = "Unable to create downloaded data file: %s"
	InvalidFileNameFormat              = "Skipped downloading file with invalid filename: %s"
	FileProcessingFailureErrorFormat   = "Failed to process downloaded file: %s (MD5: %s)"
	FileTrackingFailureErrorFormat     = "Failed to track downloaded file: %s"
	CopyFileDataResponseFailureMessage = "Failed reading file data from response body"
	FetchFileErrorMessage              = "Failed to fetch file"
	EmptyCustomerIDFailureMessage      = "empty-customer-id"
	DownloadedFilesOnDiskWarning       = "WARNING: Downloaded files are preserved for debugging purposes, but contain sensitive data. Please be sure to erase them after use."
)

//go:generate counterfeiter . urdClient
type urdClient interface {
	ListFiles(cursor string, catalogedOnOrAfter time.Time) ([]File, string, error)
	FetchFile(string) (io.ReadCloser, error)
}

//go:generate counterfeiter . processedFileTracker
type processedFileTracker interface {
	TrackFile(md5 string, catalogedAt time.Time) error
	IsFileProcessed(md5 string) bool
	MostRecentlyTrackedTimeMinusClockSkew() time.Time
}

type Visitor struct {
	logger      lager.Logger
	client      urdClient
	fileTracker processedFileTracker
}

func NewVisitor(logger lager.Logger, client urdClient, fileTracker processedFileTracker) *Visitor {
	return &Visitor{logger: logger, client: client, fileTracker: fileTracker}
}

func (v *Visitor) Visit(workingDir string, cleanupFiles bool, do func(filename string, fileMetadata Metadata) error) error {
	var cursor string
	var files []File
	var err error

	if !cleanupFiles {
		defer v.logger.Info(DownloadedFilesOnDiskWarning)
	}

	for moreFiles := true; moreFiles; moreFiles = (cursor != "") {
		files, cursor, err = v.client.ListFiles(cursor, v.fileTracker.MostRecentlyTrackedTimeMinusClockSkew())
		if err != nil {
			return errors.Wrap(err, ListFilesErrorMessage)
		}

		for _, file := range files {
			if file.Metadata.CustomerID == "" {
				v.logger.Info(EmptyCustomerIDFailureMessage)
				continue
			}

			if file.Metadata.Filename != filepath.Base(file.Metadata.Filename) {
				v.logger.Info(fmt.Sprintf(InvalidFileNameFormat, file.Metadata.Filename))
				continue
			}

			if v.fileTracker.IsFileProcessed(file.Metadata.FileMD5Checksum) {
				continue
			}

			fileContents, err := v.client.FetchFile(file.Download.URL)
			if err != nil {
				return errors.Wrap(err, FetchFileErrorMessage)
			}

			fileName := filepath.Join(workingDir, file.Metadata.FileID+"_"+file.Metadata.Filename)
			f, err := os.Create(fileName)
			if err != nil {
				return errors.Wrapf(err, CreateDataFileErrorFormat, fileName)
			}
			if cleanupFiles {
				defer func() {
					err := os.Remove(fileName)
					if err != nil {
						v.logger.Info(fmt.Sprintf(CleanupFileErrorFormat, fileName), lager.Data{"err": err.Error()})
					}
				}()
			}
			defer f.Close()

			_, err = io.Copy(f, fileContents)
			if err != nil {
				return errors.Wrap(err, CopyFileDataResponseFailureMessage)
			}

			err = do(f.Name(), file.Metadata)
			if err != nil {
				return errors.Wrapf(err, FileProcessingFailureErrorFormat, f.Name(), file.Metadata.FileMD5Checksum)
			}

			err = v.fileTracker.TrackFile(file.Metadata.FileMD5Checksum, file.Metadata.CatalogedAt)
			if err != nil {
				return errors.Wrapf(err, FileTrackingFailureErrorFormat, f.Name())
			}
		}
	}

	return nil

}
