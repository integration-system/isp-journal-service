package service

import (
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/config"
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
	assert := assert.New(t)

	initConfigForSearchWithCursor()
	request := search.SearchWithCursorRequest{
		Request:   initRequestSearchWitchCursor(),
		BatchSize: 3,
	}

	response, err := NewSearchWithCursor().Search(request)
	assert.NoError(err)
	assert.Equal(len(response.Items), 3)

	request.CursorId = response.CursorId
	response, err = NewSearchWithCursor().Search(request)
	assert.NoError(err)
	assert.Equal(len(response.Items), 1)

	request.CursorId = response.CursorId
	response, err = NewSearchWithCursor().Search(request)
	assert.NoError(err)
	assert.Equal(len(response.Items), 0)

	request.CursorId = "not found"
	response, err = NewSearchWithCursor().Search(request)
	assert.Error(err)
	assert.Nil(response)
}
