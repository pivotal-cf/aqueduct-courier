package coreconsumption

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/pivotal-cf/om/api"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	CoreCountsAPI                      = "/api/v0/download_core_consumption"
	CoreCountsRequestError             = "error retrieving core counts data"
	RequestFailureErrorFormat          = "Failed %s %s"
	RequestUnexpectedStatusErrorFormat = "%s %s returned with unexpected status %d"
	ReadResponseBodyFailureFormat      = "Unable to read response from %s"
)

type Service struct {
	Requestor Requestor
}

//go:generate counterfeiter . Requestor
type Requestor interface {
	Curl(input api.RequestServiceCurlInput) (api.RequestServiceCurlOutput, error)
}

type CoreCount struct {
	TimeReported      time.Time
	ProductIdentifier string
	PhysicalCoreCount int
	VirtualCoreCount  int
}

func (s *Service) CoreCounts() (io.Reader, error) {

	records, err := s.makeRequest(CoreCountsAPI)
	if err != nil {
		return nil, errors.Wrap(err, CoreCountsRequestError)
	}

	var counts []CoreCount
	for i := 0; i < len(records); i++ {

		coreCount, err := convertToCoreCount(records[i])
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse row in CSV with index '%d'", i)
		}

		counts = append(counts, *coreCount)
	}

	return convertToJson(counts)
}

func (s *Service) makeRequest(path string) ([][]string, error) {
	input := api.RequestServiceCurlInput{
		Path:    path,
		Method:  http.MethodGet,
		Headers: make(http.Header),
	}
	resp, err := s.Requestor.Curl(input)
	if err != nil {
		return nil, errors.Wrapf(err, RequestFailureErrorFormat, http.MethodGet, path)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf(RequestUnexpectedStatusErrorFormat, http.MethodGet, path, resp.StatusCode))
	}

	r := csv.NewReader(resp.Body)
	return r.ReadAll()
}

func convertToInt(str string, field string) (int, error) {

	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse '%s' field in CSV", field)
	}
	return i, nil
}

func convertToCoreCount(record []string) (*CoreCount, error) {
	timeInt, err := convertToInt(record[0], "collectedAt")
	if err != nil {
		return nil, err
	}

	product := record[1]

	countOne, err := convertToInt(record[2], "contOne")
	if err != nil {
		return nil, err
	}

	countTwo, err := convertToInt(record[3], "countTwo")
	if err != nil {
		return nil, err
	}

	return &CoreCount{
		TimeReported:      time.Unix(int64(timeInt), 0).UTC(),
		ProductIdentifier: product,
		PhysicalCoreCount: countOne,
		VirtualCoreCount:  countTwo,
	}, nil
}

func convertToJson(counts []CoreCount) (io.Reader, error) {

	jsonBytes, err := json.Marshal(counts)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to convert to JSON")
	}

	return bytes.NewReader(jsonBytes), nil
}
