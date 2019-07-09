package utils

import (
	"compress/gzip"
	"encoding/csv"
	"github.com/integration-system/isp-lib/logger"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetTempFilePath() (string, error) {
	if temp, err := ioutil.TempDir("", ""); err != nil {
		return "", err
	} else {
		return filepath.Join(temp, "info"), nil
	}
}

func UseCsvWriter(path string, csvSep rune, f func(writer *csv.Writer) error) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(file)
	csvWriter := csv.NewWriter(gzipWriter)
	csvWriter.Comma = csvSep
	defer func() {
		if csvWriter != nil {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				logger.Error(err)
			}
		}
		if gzipWriter != nil {
			if err := gzipWriter.Flush(); err != nil {
				logger.Error(err)
			}
			if err := gzipWriter.Close(); err != nil {
				logger.Error(err)
			}
		}
		if file != nil {
			if err = file.Close(); err != nil {
				logger.Error(err)
			}
		}
	}()
	return f(csvWriter)
}
