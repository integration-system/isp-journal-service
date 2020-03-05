package controller

import (
	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/v2/streaming"

	"google.golang.org/grpc/metadata"
	"io"
	"isp-journal-service/service"
)

var (
	LogController = logImpl{}
)

type logImpl struct {
}

func (logImpl) Transfer(stream streaming.DuplexMessageStream, md metadata.MD) error {
	_, err := streaming.ReadFile(stream, func(bf streaming.BeginFile) (io.WriteCloser, error) {
		info, err := transfer.GetLogInfo(bf)
		if err != nil {
			return nil, err
		}

		return service.LogService.OpenWriter(*info)
	}, true)

	return err
}
