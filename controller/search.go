package controller

import (
	"github.com/integration-system/isp-journal/search"
	"isp-journal-service/service"
	"isp-journal-service/shared"
)

var (
	SearchController = searchImpl{}
)

type searchImpl struct{}

func (searchImpl) Search(request search.SearchRequest) ([]shared.SearchResponse, error) {
	return service.NewSearchService().Search(request)
}
