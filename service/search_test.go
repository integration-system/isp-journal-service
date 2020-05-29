//nolint
package service

import (
	"testing"
	"time"

	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/stretchr/testify/assert"
	"isp-journal-service/conf"
)

func TestSearchImpl_Search(t *testing.T) {
	a := assert.New(t)

	initConfigForSearch()
	req, err := initRequestSearch()
	a.NoError(err)

	resp, err := NewSearchService().Search(req)
	a.NoError(err)
	a.Equal(2, len(resp))
}

func initRequestSearch() (*search.SearchRequest, error) {
	from, err := time.Parse("2006-01-02T15:04:05.999-07:00", "2019-06-10T08:10:51.964-00:00")
	if err != nil {
		return nil, err
	}
	to, err := time.Parse("2006-01-02T15:04:05.999-07:00", "2019-06-10T14:29:51.964-00:00")
	if err != nil {
		return nil, err
	}
	return &search.SearchRequest{
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

func initConfigForSearch() {
	config.UnsafeSetRemote(&conf.RemoteConfig{BaseLogDirectory: "./test"})
}
