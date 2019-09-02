package controller

import (
	"github.com/integration-system/isp-journal/search"
	"isp-journal-service/service"
)

var (
	SearchController = searchImpl{}
)

type searchImpl struct{}

func (searchImpl) Search(request search.SearchRequest) ([]search.SearchResponse, error) {
	return service.NewSearchService().Search(request)
}

func (searchImpl) SearchWithCursor(request search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	return service.NewSearchWithCursor().Search(request)
}
