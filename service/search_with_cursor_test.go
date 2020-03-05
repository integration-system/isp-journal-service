//nolint
package service

import (
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/stretchr/testify/assert"
	"isp-journal-service/conf"
	"testing"
)

func initRequestSearchWitchCursor() search.SearchRequest {
	return search.SearchRequest{
		ModuleName: "mdm-test-service",
	}
}

func initConfigForSearchWithCursor() {
	config.UnsafeSetRemote(&conf.RemoteConfig{BaseLogDirectory: "./test", CursorLifetime: 10})
}

func TestCursorService_Search(t *testing.T) {
	a := assert.New(t)

	initConfigForSearchWithCursor()
	request := &search.SearchWithCursorRequest{
		Request:   initRequestSearchWitchCursor(),
		BatchSize: 3,
	}

	response, err := CursorService.Search(request)
	a.NoError(err)
	a.Equal(len(response.Items), 3)

	request.CursorId = response.CursorId
	response, err = CursorService.Search(request)
	a.NoError(err)
	a.Equal(len(response.Items), 1)

	request.CursorId = response.CursorId
	response, err = CursorService.Search(request)
	a.NoError(err)
	a.Equal(len(response.Items), 0)

	request.CursorId = "not found"
	response, err = CursorService.Search(request)
	a.Error(err)
	a.Nil(response)
}
