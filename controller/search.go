package controller

import (
	"isp-journal-service/service"
	"isp-journal-service/shared"
)

var (
	SearchController = searchImpl{}
)

type searchImpl struct{}

func (searchImpl) Search(request shared.SearchRequest) ([]shared.SearchResponse, error) {
	return service.NewSearchService().Search(request)
}
