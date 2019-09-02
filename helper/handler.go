package helper

import (
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/streaming"
	"isp-journal-service/controller"
)

type logHandler struct {
	Transfer streaming.StreamConsumer `method:"transfer" group:"log" inner:"true"`
}

type searchHandler struct {
	Search           func(search.SearchRequest) ([]search.SearchResponse, error)                    `method:"search" group:"log" inner:"true"`
	SearchWithCursor func(search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) `method:"search_with_cursor" group:"log" inner:"true"`
}

type exportHandler struct {
	Export streaming.StreamConsumer `method:"export" group:"log" inner:"true"`
}

func GetAllHandlers() []interface{} {
	return []interface{}{
		&logHandler{
			Transfer: controller.LogController.Transfer,
		},
		&searchHandler{
			Search:           controller.SearchController.Search,
			SearchWithCursor: controller.SearchController.SearchWithCursor,
		},
		&exportHandler{
			Export: controller.ExportController.Export,
		},
	}
}
