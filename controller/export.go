package controller

import (
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/resources"
	"github.com/integration-system/isp-lib/streaming"
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/metadata"
	"isp-journal-service/service"
	"isp-journal-service/shared"
	"os"
	"path/filepath"
)

var ExportController = exportImpl{}

type exportImpl struct{}

func (exportImpl) Export(stream streaming.DuplexMessageStream, md metadata.MD) error {
	filePath, err := resources.GetTempFilePath()
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(filepath.Dir(filePath)) }()

	request := new(shared.SearchRequest)
	message, err := stream.Recv()
	if err != nil {
		return err
	}
	body := backend.ResolveBody(message)
	if err := utils.ConvertGrpcToGo(body, request); err != nil {
		return err
	}

	if err := service.NewImportService(*request).Export(filePath); err != nil {
		return err
	}

	bf := streaming.BeginFile{
		FileName:     "log.csv.gz",
		ContentType:  "log/csv",
		FormDataName: "log",
	}
	if err := streaming.WriteFile(stream, filePath, bf); err != nil {
		return err
	}
	return nil
}