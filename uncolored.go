package main

import (
	"github.com/reconquest/loreley"
	"io"
)

type uncoloredWriter struct {
	writer io.Writer
}

func (writer uncoloredWriter) Write(data []byte) (int, error)  {
	return writer.writer.Write([]byte(loreley.TrimStyles(string(data))))
}
