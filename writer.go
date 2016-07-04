package main

import (
	"io"
	"regexp"

	"github.com/reconquest/loreley"
)

var (
	reLogPrefix = regexp.MustCompile(
		`^([\d-]+\s+[\d:]+)\s+\[\w+\]\s+\[[\w\#\d]+\] (.*)`, 
	)
)

type uncolored struct {
	writer io.Writer
}

type unprefixed struct {
	writer io.Writer
}

func (writer unprefixed) Write(data []byte) (int, error) {
	return writer.writer.Write(
		reLogPrefix.ReplaceAll(
			data,
			[]byte(`$2`),
		),
	)
}

func (writer uncolored) Write(data []byte) (int, error) {
	return writer.writer.Write([]byte(loreley.TrimStyles(string(data))))
}
