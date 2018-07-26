package data

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

const (
	OpsManagerProductType = "ops_manager"
	DirectorProductType   = "p-bosh"

	ResourcesDataType        = "resources"
	VmTypesDataType          = "vm_types"
	DiagnosticReportDataType = "diagnostic_report"
	DeployedProductsDataType = "deployed_products"
	InstallationsDataType    = "installations"
	PropertiesDataType       = "properties"

	MetadataFileName = "aqueduct_metadata"

	ReadMetadataFileError             = "Unable to read metadata file"
	InvalidMetadataFileError          = "Metadata file is invalid"
	ExtraFilesInTarMessageFormat      = "Tar file %s contains unexpected extra files"
	MissingFilesInTarMessageFormat    = "Tar file %s is missing contents"
	InvalidFilesInTarMessageFormat    = "Tar file %s content does not match recorded value"
	InvalidFileNameInTarMessageFormat = "Tar file %s has files with invalid names"
	UnableToListFilesMessageFormat    = "Unable to list files in %s"
)

type Metadata struct {
	EnvType      string
	CollectedAt  string
	CollectionId string
	FileDigests  []FileDigest
}
type FileDigest struct {
	Name        string
	MimeType    string
	MD5Checksum string
	ProductType string
	DataType    string
}
type FileValidator struct {
	tReader tarReader
}

//go:generate counterfeiter . tarReader
type tarReader interface {
	ReadFile(fileName string) ([]byte, error)
	FileMd5s() (map[string]string, error)
	TarFilePath() string
}

func NewFileValidator(tReader tarReader) *FileValidator {
	return &FileValidator{tReader: tReader}
}

func (v *FileValidator) Validate() error {
	metadata, err := v.readMetadata()
	if err != nil {
		return err
	}

	fileMd5s, err := v.tReader.FileMd5s()
	if err != nil {
		return errors.Wrapf(err, UnableToListFilesMessageFormat)
	}
	delete(fileMd5s, MetadataFileName)

	for _, digest := range metadata.FileDigests {
		if strings.Contains(digest.Name, ".") || strings.Contains(digest.Name, "/") {
			return errors.Errorf(InvalidFileNameInTarMessageFormat, v.tReader.TarFilePath())
		}
		if checksum, exists := fileMd5s[digest.Name]; exists {
			if digest.MD5Checksum != checksum {
				return errors.Errorf(InvalidFilesInTarMessageFormat, v.tReader.TarFilePath())
			}
			delete(fileMd5s, digest.Name)
		} else {
			return errors.Errorf(MissingFilesInTarMessageFormat, v.tReader.TarFilePath())
		}
	}
	if len(fileMd5s) != 0 {
		return errors.Errorf(ExtraFilesInTarMessageFormat, v.tReader.TarFilePath())
	}

	return nil
}

func (v *FileValidator) readMetadata() (Metadata, error) {
	var metadata Metadata

	metadataBytes, err := v.tReader.ReadFile(MetadataFileName)
	if err != nil {
		return metadata, errors.Wrap(err, ReadMetadataFileError)
	}

	decoder := json.NewDecoder(bytes.NewReader(metadataBytes))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&metadata)
	if err != nil {
		return metadata, errors.Wrapf(err, InvalidMetadataFileError)
	}

	return metadata, nil
}
