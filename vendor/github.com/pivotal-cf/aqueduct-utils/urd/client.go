package urd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
)

const (
	WouQueriesPath = "/api/v1/queries"

	QueryRequestCreationFailureMessage = "failed to create request object for file list query"
	QueryRequestFailureMessage         = "failed to perform file list request"
	ListFilesResponseCodeErrorFormat   = "unexpected response code %d attempting to list files"
	ReadQueryResultFailureMessage      = "failed reading query result"
	FetchRequestCreationFailureMessage = "failed to create request object for file fetch"
	FetchRequestFailureMessage         = "failed to perform file fetch"
	FetchFileResponseCodeErrorFormat   = "unexpected response code %d attempting to fetch file"
)

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}


type DownloadInfo struct {
	URL string `json:"url"`
}

type File struct {
	Metadata Metadata     `json:"metadata"`
	Download DownloadInfo `json:"download"`
}

type fileQueryResult struct {
	Files  []File `json:"files"`
	Cursor string `json:"nextPageCursor"`
}

type Client struct {
	logger    lager.Logger
	client    httpClient
	url       string
	dataSetId string
}

func NewClient(logger lager.Logger, client httpClient, url, dataSetId string) *Client {
	return &Client{
		logger:    logger,
		client:    client,
		url:       url,
		dataSetId: dataSetId,
	}
}

func (uc *Client) ListFiles(startingCursor string, catalogedOnOrAfter time.Time) ([]File, string, error) {
	log := uc.logger.Session("list-files")
	log.Info("start")
	defer log.Info("finish")

	reqBody, err := urdJson(uc.dataSetId, startingCursor, catalogedOnOrAfter)
	if err != nil {
		return []File{}, "", errors.Wrap(err, QueryRequestCreationFailureMessage)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		uc.url+WouQueriesPath,
		reqBody,
	)
	if err != nil {
		return []File{}, "", errors.Wrap(err, QueryRequestCreationFailureMessage)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := uc.client.Do(req)
	if err != nil {
		return []File{}, "", errors.Wrap(err, QueryRequestFailureMessage)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []File{}, "", errors.Errorf(ListFilesResponseCodeErrorFormat, resp.StatusCode)
	}

	var result fileQueryResult
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return []File{}, "", errors.Wrap(err, ReadQueryResultFailureMessage)
	}

	return result.Files, result.Cursor, nil
}

func (uc *Client) FetchFile(fileURL string) (io.ReadCloser, error) {
	log := uc.logger.Session("fetch-file")
	log.Info("start")
	defer log.Info("finish")

	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, FetchRequestCreationFailureMessage)
	}

	resp, err := uc.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, FetchRequestFailureMessage)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(FetchFileResponseCodeErrorFormat, resp.StatusCode)
	}

	return resp.Body, nil
}

func urdJson(dataSetId, cursor string, catalogedOnOrAfter time.Time) (io.Reader, error) {
	urdMap := map[string]interface{}{
		"limit":              10,
		"dataSetId":          dataSetId,
		"catalogedOnOrAfter": catalogedOnOrAfter.Format(time.RFC3339),
	}
	if cursor != "" {
		urdMap["cursor"] = cursor
	}
	urdJson, err := json.Marshal(urdMap)

	return bytes.NewReader(urdJson), err
}
