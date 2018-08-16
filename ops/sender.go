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
	"io/ioutil"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/collections/foundation"
	TarMimeType            = "application/tar"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	ReadMetadataFileError         = "Unable to read metadata file"
	ReadDataFileError             = "Unable to read data file"
	InvalidMetadataFileError      = "Metadata file is invalid"
	UnauthorizedErrorMessage      = "User is not authorized to perform this action"
	UnexpectedServerErrorFormat   = "There was an issue sending data. Please try again or contact your Pivotal field team if this error persists. Error Code %s"

	FileValidationFailedMessageFormat = "File %s is invalid"
)

type SendExecutor struct{}

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate counterfeiter . tarReader
type tarReader interface {
	ReadFile(string) ([]byte, error)
	FileMd5s() (map[string]string, error)
}

//go:generate counterfeiter . validator
type validator interface {
	Validate() error
}

func (s SendExecutor) Send(client httpClient, reader tarReader, tValidator validator, tarFilePath, dataLoaderURL, apiToken, senderVersion string) error {
	metadataContent, err := reader.ReadFile(data.MetadataFileName)
	if err != nil {
		return errors.Wrap(err, ReadMetadataFileError)
	}

	var metadata data.Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return errors.Wrap(err, InvalidMetadataFileError)
	}

	if err := tValidator.Validate(); err != nil {
		return errors.Wrapf(err, FileValidationFailedMessageFormat, tarFilePath)
	}

	req, err := makeFileUploadRequest(
		tarFilePath,
		apiToken,
		dataLoaderURL+PostPath,
		senderVersion,
		metadata,
	)
	if err != nil {
		return errors.Wrap(err, RequestCreationFailureMessage)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, PostFailedMessage)
	}
	defer resp.Body.Close()

	switch statusCode := resp.StatusCode; statusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errors.New(UnauthorizedErrorMessage)
	default:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf(UnexpectedServerErrorFormat, "unknown")
		}

		var errResp map[string]map[string]string
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return errors.Errorf(UnexpectedServerErrorFormat, "unknown")
		}
		return errors.Errorf(UnexpectedServerErrorFormat, errResp["error"]["uuid"])
	}
}

func constructFileMetadataReader(metadata data.Metadata, fileName, senderVersion string, hashWriter hash.Hash) (io.Reader, error) {
	metadataMap := map[string]interface{}{
		"filename":        fileName,
		"fileContentType": TarMimeType,
		"fileMd5Checksum": base64.StdEncoding.EncodeToString(hashWriter.Sum([]byte{})),
		"collectedAt":     metadata.CollectedAt,
		"customMetadata": map[string]string{
			"SenderVersion": senderVersion,
			"EnvType":       metadata.EnvType,
			"CollectionId":  metadata.CollectionId,
		},
	}
	metadataJson, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(metadataJson), nil
}

func makeFileUploadRequest(filePath, apiToken, uploadURL, senderVersion string, metadata data.Metadata) (*http.Request, error) {
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

	metadataReader, err := constructFileMetadataReader(metadata, filepath.Base(filePath), senderVersion, hashWriter)
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
