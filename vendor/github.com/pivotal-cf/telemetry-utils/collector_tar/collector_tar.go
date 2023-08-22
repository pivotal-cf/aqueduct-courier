package collector_tar

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

const (
	OpsManagerProductType = "ops_manager"
	DirectorProductType   = "p-bosh"

	ResourcesDataType              = "resources"
	VmTypesDataType                = "vm_types"
	DiagnosticReportDataType       = "diagnostic_report"
	DeployedProductsDataType       = "deployed_products"
	InstallationsDataType          = "installations"
	PropertiesDataType             = "properties"
	CertificatesDataType           = "certificates"
	CertificateAuthoritiesDataType = "certificate_authorities"
	PendingChangesDataType         = "product_changes"
	AppUsageDataType               = "app_usage"
	ServiceUsageDataType           = "service_usage"
	TaskUsageDataType              = "task_usage"
	CoreCountsDataType             = "core_counts"

	OpsManagerCollectorDataSetId      = "opsmanager"
	UsageServiceCollectorDataSetId    = "usage_service"
	CoreConsumptionCollectorDataSetId = "core_consumption"

	MetadataFileName = "metadata"

	ReadMetadataFileError            = "Unable to read metadata file"
	InvalidMetadataFileError         = "Metadata file is invalid"
	ExtraFilesInTarMessageError      = "Tar contains unexpected extra files"
	MissingFilesInTarMessageError    = "Tar is missing contents"
	InvalidFilesInTarMessageError    = "Tar content does not match recorded value"
	InvalidFileNameInTarMessageError = "Tar has files with invalid names"
	UnableToListFilesMessageError    = "Unable to list files in tar"
)

type Metadata struct {
	EnvType            string
	CollectedAt        string
	CollectionId       string
	FoundationId       string
	FoundationNickname string
	FileDigests        []FileDigest
	CollectorVersion   string
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
		return errors.Wrapf(err, UnableToListFilesMessageError)
	}
	delete(fileMd5s, MetadataFileName)

	for _, digest := range metadata.FileDigests {
		if strings.Contains(digest.Name, ".") || strings.Contains(digest.Name, "/") {
			return errors.New(InvalidFileNameInTarMessageError)
		}
		if checksum, exists := fileMd5s[digest.Name]; exists {
			if digest.MD5Checksum != checksum {
				return errors.New(InvalidFilesInTarMessageError)
			}
			delete(fileMd5s, digest.Name)
		} else {
			return errors.New(MissingFilesInTarMessageError)
		}
	}
	if len(fileMd5s) != 0 {
		return errors.New(ExtraFilesInTarMessageError)
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
