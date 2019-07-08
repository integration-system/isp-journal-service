package helper

import (
	"github.com/integration-system/isp-lib/streaming"
	"isp-journal-service/controller"
)

type logHandler struct {
	Transfer streaming.StreamConsumer `method:"transfer" group:"log" inner:"true"`
}

func GetAllHandlers() []interface{} {
	return []interface{}{
		&logHandler{
			Transfer: controller.LogController.Transfer,
		},
	}
}
