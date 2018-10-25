package urd

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

const (
	ClockSkewAllowance = 3 * time.Minute
	bufferTime = 2 * ClockSkewAllowance
)

type FileTracker struct {
	RejectMD5s      []string
	MD5s            map[string]time.Time
	LatestTimestamp time.Time
	filePath        string
	rejectMD5Map    map[string]bool
}

func NewFileTracker(filePath string) (*FileTracker, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	tracker := FileTracker{MD5s: map[string]time.Time{}, filePath: filePath, rejectMD5Map: map[string]bool{}}
	err = json.Unmarshal(content, &tracker)
	for _, reject := range tracker.RejectMD5s {
		tracker.rejectMD5Map[reject] = true
	}
	return &tracker, err
}

func (ft *FileTracker) TrackFile(md5 string, timestamp time.Time) error {
	if timestamp.After(ft.LatestTimestamp) {
		ft.LatestTimestamp = timestamp
	} else if timestamp.Add(bufferTime).Before(ft.LatestTimestamp) {
		return nil
	}
	for trackedMD5, trackedTimestamp := range ft.MD5s {
		if trackedTimestamp.Add(bufferTime).Before(ft.LatestTimestamp) {
			delete(ft.MD5s, trackedMD5)
		}
	}
	ft.MD5s[md5] = timestamp
	data, _ := json.Marshal(ft)
	return ioutil.WriteFile(ft.filePath, data, 0644)
}

func (ft *FileTracker) IsFileProcessed(md5 string) bool {
	_, inMD5s := ft.MD5s[md5]
	if inMD5s {
		return true
	}

	return ft.rejectMD5Map[md5]
}

func (ft *FileTracker) MostRecentlyTrackedTimeMinusClockSkew() time.Time {
	adjustedTime := ft.LatestTimestamp.Add(-ClockSkewAllowance)
	if adjustedTime.Before(time.Time{}) {
		return time.Time{}
	}
	return adjustedTime
}
