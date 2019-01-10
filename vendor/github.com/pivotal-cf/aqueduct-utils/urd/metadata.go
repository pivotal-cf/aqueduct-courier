package urd

import "time"

type Metadata struct {
	DataLoaderName    string                 `json:"dataLoaderName"`
	DataLoaderVersion string                 `json:"dataLoaderVersion"`
	DataSetID         string                 `json:"dataSetId"`
	CollectedAt       string                 `json:"collectedAt"`
	FileContentType   string                 `json:"fileContentType"`
	FileMD5Checksum   string                 `json:"fileMd5Checksum"`
	FileID            string                 `json:"fileId"`
	Filename          string                 `json:"filename"`
	CustomerID        string                 `json:"customerId"`
	CatalogedAt       time.Time              `json:"catalogedAt"`
	ReceivedAt        string                 `json:"receivedAt"`
	CustomMetadata    map[string]interface{} `json:"customMetadata"`
}

