package main

import (
	"encoding/json"
	"os"
	"sync"
)

type Metrics struct {
	HostID    string `json:"hostid"`
	Tag       string `json:"tag"`
	Body      string `json:"body"`
	Timestamp int64  `json:"timestamp`
}

type MetricsWriter struct {
	file *os.File
	lock sync.Mutex
}

func NewMetricsWriter(logFile string) (*MetricsWriter, error) {
	var err error
	mw := &MetricsWriter{}

	mw.file, err = os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return mw, nil
}

func (mw *MetricsWriter) Write(metrics Metrics) error {
	js, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	// jsonlなので、最後に改行を追加
	js = append(js, '\n')

	mw.lock.Lock()
	defer mw.lock.Unlock()

	_, err = mw.file.Write(js)
	if err != nil {
		return err
	}
	return nil
}
