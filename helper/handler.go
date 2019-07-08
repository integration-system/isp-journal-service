package helper

import (
	"github.com/integration-system/isp-lib/streaming"
	"isp-journal-service/controller"
	"isp-journal-service/shared"
)

type logHandler struct {
	Transfer streaming.StreamConsumer `method:"transfer" group:"log" inner:"true"`
}

type searchHandler struct {
	Search func(shared.SearchRequest) ([]shared.SearchResponse, error) `method:"search" group:"log" inner:"true"`
}

func GetAllHandlers() []interface{} {
	return []interface{}{
		&logHandler{
			Transfer: controller.LogController.Transfer,
		},
		&searchHandler{
			Search: controller.SearchController.Search,
		},
	}
}
