package controller

import (
	"io"

	"github.com/integration-system/isp-journal/transfer"
	"github.com/integration-system/isp-lib/v2/streaming"
	log "github.com/integration-system/isp-log"
	"isp-journal-service/log_code"

	"google.golang.org/grpc/metadata"
	"isp-journal-service/service"
)

var (
	LogController = logImpl{}
)

type logImpl struct {
}

func (logImpl) Transfer(stream streaming.DuplexMessageStream, md metadata.MD) error {
	var dir string
	_, err := streaming.ReadFile(stream, func(bf streaming.BeginFile) (io.WriteCloser, error) {
		info, err := transfer.GetLogInfo(bf)
		if err != nil {
			return nil, err
		}

		var writeCloser io.WriteCloser
		writeCloser, dir, err = service.LogService.OpenWriter(*info)
		return writeCloser, err
	}, true)
	if err != nil {
		return err
	}
	go func() {
		err := service.NewElasticPublisher().Publish(dir)
		if err != nil {
			log.Error(log_code.ErrorElastic, err)
		}
	}()
	return nil
}
