package service

import (
	"github.com/integration-system/isp-lib/config"
	"github.com/stretchr/testify/assert"
	"isp-journal-service/conf"
	"isp-journal-service/shared"
	"testing"
	"time"
)

func TestSearchImpl_Search(t *testing.T) {
	assert := assert.New(t)

	initConfig()

	req, err := initRequestSearch()
	assert.NoError(err)

	resp, err := SearchService.Search(*req)
	assert.NoError(err)
	assert.NotNil(resp)
}

/*func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	initConfig()

	bytes, err := entry.MarshalToBytes(&entry.Entry{
		ModuleName: "mdm-test-service",
		Host:       "127.0.0.1",
		Event:      "OLD_RECORD",
		Level:      "INFO",
		Time:       "2019-06-10",
		Request:    nil,
		Response:   nil,
		ErrorText:  "",
	})
	assert.NoError(err)

	path := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	file, err := os.Create(path + "/2019-06-10/mdm-test-service/127.0.0.1__2019-06-10T08-25-33.930.log.gz")
	assert.NoError(err)
	defer func() { _ = file.Close() }()

	gzipWriter := gzip.NewWriter(file)
	defer func() {
		_ = gzipWriter.Flush()
		_ = gzipWriter.Close()
	}()

	_, err = gzipWriter.Write(bytes)
	assert.NoError(err)
}*/

func initRequestSearch() (*shared.SearchRequest, error) {
	from, err := time.Parse("2006-01-02", "2019-06-10")
	if err != nil {
		return nil, err
	}
	to, err := time.Parse("2006-01-02", "2019-06-10")
	if err != nil {
		return nil, err
	}
	return &shared.SearchRequest{
		ModuleName: "mdm-test-service",
		From:       from,
		To:         to,
		Host:       []string{"127.0.0.1"},
		Event:      []string{"NEW_RECORD"},
		Level:      []string{"INFO"},
		Limit:      4,
		Offset:     0,
	}, nil
}

func initConfig() {
	config.UnsafeSetRemote(&conf.RemoteConfig{BaseLogDirectory: "./test"})
}
