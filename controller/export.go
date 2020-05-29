package controller

import (
	"os"
	"path/filepath"

	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/resources"
	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/utils"
	"google.golang.org/grpc/metadata"
	"isp-journal-service/service"
)

var ExportController = exportImpl{}

type exportImpl struct{}

func (exportImpl) Export(stream streaming.DuplexMessageStream, md metadata.MD) error {
	filePath, err := resources.GetTempFilePath()
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(filepath.Dir(filePath)) }()

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

	err = service.NewImportService(request).Export(filePath)
	if err != nil {
		return err
	}

	bf := streaming.BeginFile{
		FileName:     "log.csv.gz",
		ContentType:  "log/csv",
		FormDataName: "log",
	}
	err = streaming.WriteFile(stream, filePath, bf)
	if err != nil {
		return err
	}
	return nil
}
