package controller

import (
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/metadata"
	"isp-journal-service/service"
)

var ExportController = exportImpl{}

type exportImpl struct{}

func (exportImpl) Export(stream streaming.DuplexMessageStream, md metadata.MD) error {
	bf := streaming.BeginFile{
		FileName:     "log.csv.gz",
		ContentType:  "log/csv",
		FormDataName: "log",
	}

	writer, err := streaming.NewFileStreamWriter(stream, bf)
	if err != nil {
		return err
	}

	request := new(search.SearchRequest)
	message, err := stream.Recv()
	if err != nil {
		return err
	}
	body := backend.ResolveBody(message)
	err = utils.ConvertGrpcToGo(body, request)
	if err != nil {
		return err
	}

	err = service.NewImportService(request).Export(writer)
	if err != nil {
		return err
	}

	return nil
}
