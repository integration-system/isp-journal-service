package service

import (
	"compress/gzip"
	"fmt"
	io2 "github.com/integration-system/isp-io"
	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/config"
	"io"
	"isp-journal-service/conf"
	"os"
	"path/filepath"
	"time"
)

const (
	dateDirFormat  = "2006-01-02"
	timeFileFormat = "2006-01-02T15-04-05.000"
)

var (
	LogService = logService{}
)

type logService struct {
}

func (logService) OpenWriter(info transfer.LogInfo) (io.WriteCloser, error) {
	dir, err := ensureDirectory(info.ModuleName, info.CreatedAt)
	if err != nil {
		return nil, err
	}

	name := filepath.Join(dir, getFileName(info))
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	pipe := io2.NewWritePipe(file)

	if !info.Compressed {
		gzWr := gzip.NewWriter(pipe.Last())
		pipe.Unshift(gzWr)
	}

	return pipe, nil
}

func getFileName(info transfer.LogInfo) string {
	return fmt.Sprintf("%s__%s.log.gz", info.Host, info.CreatedAt.Format(timeFileFormat))
}

func ensureDirectory(moduleName string, createdAt time.Time) (string, error) {
	base := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	path := filepath.Join(base, createdAt.Format(dateDirFormat), moduleName)

	err := os.MkdirAll(path, 0755)

	return path, err
}
