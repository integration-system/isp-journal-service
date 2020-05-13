package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/v2/config"
	jsoniter "github.com/json-iterator/go"
	"isp-journal-service/conf"
)

var json = jsoniter.ConfigFastest

func getFileName(info transfer.LogInfo) string {
	return fmt.Sprintf("%s__%s.log.gz", info.Host, info.CreatedAt.Format(timeFileFormat))
}

func ensureDirectory(moduleName string, createdAt time.Time) (string, error) {
	base := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	path := filepath.Join(base, createdAt.Format(dateDirFormat), moduleName)

	err := os.MkdirAll(path, 0755)

	return path, err
}
