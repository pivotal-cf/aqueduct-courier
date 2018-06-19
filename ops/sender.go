package ops

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"hash"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/collections/foundation"
	TarMimeType            = "application/tar"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	UnexpectedResponseCodeFormat  = "Unexpected response code %d, request failed"
	ReadMetadataFileError         = "Unable to read metadata file"
	ReadDataFileError             = "Unable to read data file "
	InvalidMetadataFileError      = "Metadata file is invalid"

	ExtraFilesInTarMessageFormat   = "Tar file %s contains unexpected extra files"
	MissingFilesInTarMessageFormat = "Tar file %s is missing contents"
	InvalidFilesInTarMessageFormat = "Tar file %s content does not match recorded value"
	UnableToListFilesMessageFormat = "Unable to list files in %s"
)

type SendExecutor struct{}

//go:generate counterfeiter . tarReader
type tarReader interface {
	TarFilePath() string
	ReadFile(string) ([]byte, error)
	FileMd5s() (map[string]string, error)
}

func (s SendExecutor) Send(reader tarReader, dataLoaderURL, apiToken string) error {
	metadataContent, err := reader.ReadFile(MetadataFileName)
	if err != nil {
		return errors.Wrap(err, ReadMetadataFileError)
	}

	var metadata data.Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return errors.Wrap(err, InvalidMetadataFileError)
	}

	if err := validateTarFile(reader, metadata); err != nil {
		return err
	}

	req, err := makeFileUploadRequest(
		reader.TarFilePath(),
		apiToken,
		dataLoaderURL+PostPath,
		metadata,
	)
	if err != nil {
		return errors.Wrap(err, RequestCreationFailureMessage)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, PostFailedMessage)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf(UnexpectedResponseCodeFormat, resp.StatusCode)
	}
	return nil
}

func constructFileMetadataReader(metadata data.Metadata, fileName string, hashWriter hash.Hash) (io.Reader, error) {
	metadataMap := map[string]string{
		"filename":        fileName,
		"fileContentType": TarMimeType,
		"fileMd5Checksum": base64.StdEncoding.EncodeToString(hashWriter.Sum([]byte{})),
		"collectedAt":     metadata.CollectedAt,
		"envType":         metadata.EnvType,
		"collectionId":    metadata.CollectionId,
	}
	metadataJson, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(metadataJson), nil
}

func makeFileUploadRequest(filePath, apiToken, uploadURL string, metadata data.Metadata) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, ReadDataFileError)
	}
	defer file.Close()

	dataPart, err := writer.CreateFormFile("data", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	hashWriter := md5.New()
	if _, err := io.Copy(io.MultiWriter(dataPart, hashWriter), file); err != nil {
		return nil, err
	}

	metadataReader, err := constructFileMetadataReader(metadata, filepath.Base(filePath), hashWriter)
	if err != nil {
		return nil, err
	}

	metadataPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(metadataPart, metadataReader)
	if err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(AuthorizationHeaderKey, "Token "+apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func validateTarFile(reader tarReader, metadata data.Metadata) error {
	fileMd5s, err := reader.FileMd5s()
	if err != nil {
		return errors.Wrapf(err, UnableToListFilesMessageFormat, reader.TarFilePath())
	}

	delete(fileMd5s, MetadataFileName)

	for _, digest := range metadata.FileDigests {
		if checksum, exists := fileMd5s[digest.Name]; exists {
			if digest.MD5Checksum != checksum {
				return errors.Errorf(InvalidFilesInTarMessageFormat, reader.TarFilePath())
			}
			delete(fileMd5s, digest.Name)
		} else {
			return errors.Errorf(MissingFilesInTarMessageFormat, reader.TarFilePath())
		}
	}
	if len(fileMd5s) != 0 {
		return errors.Errorf(ExtraFilesInTarMessageFormat, reader.TarFilePath())
	}

	return nil
}
