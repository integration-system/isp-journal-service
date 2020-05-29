package service

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	io2 "github.com/integration-system/isp-io"
	"github.com/integration-system/isp-journal/transfer"
)

const (
	dateDirFormat  = "2006-01-02"
	timeFileFormat = "2006-01-02T15-04-05.000"
)

var LogService = logService{}

type logService struct{}

func (logService) OpenWriter(info transfer.LogInfo) (io.WriteCloser, string, error) {
	dir, err := ensureDirectory(info.ModuleName, info.CreatedAt)
	if err != nil {
		return nil, "", err
	}

	name := filepath.Join(dir, getFileName(info))
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, "", err
	}

	pipe := io2.NewWritePipe(file)

	if !info.Compressed {
		gzWr := gzip.NewWriter(pipe.Last())
		pipe.Unshift(gzWr)
	}

	return pipe, name, nil
}
