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

	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/placeholder"
	TarMimeType            = "application/tar"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	UnexpectedResponseCodeFormat  = "Unexpected response code %d, request failed"
	ReadMetadataFileError         = "Unable to read metadata file"
	ReadDataFileError             = "Unable to read data file "
	InvalidMetadataFileError      = "Metadata file is invalid"
)

type SendExecutor struct{}

//go:generate counterfeiter . tarReader
type tarReader interface {
	ReadFile(string) ([]byte, error)
	TarFilePath() string
}

func (s SendExecutor) Send(reader tarReader, dataLoaderURL, apiToken string) error {
	metadataContent, err := reader.ReadFile(MetadataFileName)
	if err != nil {
		return errors.Wrap(err, ReadMetadataFileError)
	}

	var metadata Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return errors.Wrap(err, InvalidMetadataFileError)
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
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf(UnexpectedResponseCodeFormat, resp.StatusCode)
	}
	return nil
}

func constructFileMetadataReader(metadata Metadata, fileName string, hashWriter hash.Hash) (io.Reader, error) {
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

func makeFileUploadRequest(filePath, apiToken, uploadURL string, metadata Metadata) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, ReadDataFileError)
	}

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
