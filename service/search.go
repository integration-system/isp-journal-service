package service

import (
	"compress/gzip"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"isp-journal-service/conf"
	"isp-journal-service/shared"
	"os"
	"time"
)

const dirLayout = "2006-01-02"

var (
	SearchService = &searchService{}
)

type searchService struct {
	count    int
	offset   int
	response []shared.SearchResponse
}

func (s *searchService) Search(request shared.SearchRequest) ([]shared.SearchResponse, error) {
	s.count = 0
	s.offset = 0
	s.response = make([]shared.SearchResponse, 0, request.Limit)

	if request.From.IsZero() {
		if !request.To.IsZero() {
			return nil, status.Error(codes.InvalidArgument, "expected FROM if specified TO")
		}
		request.From = time.Now().UTC()
	}

	if !request.To.IsZero() {
		if request.To.Before(request.From) {
			return nil, status.Error(codes.InvalidArgument, "expected FROM will before TO")
		}
	} else {
		request.To = time.Now().UTC()
	}

	if files, err := s.getFiles(request); err != nil {
		return nil, err
	} else {
		if err := s.readFiles(request, files); err != nil {
			return nil, err
		}
	}
	for _, value := range s.response {
		logger.Info(value)
	}
	if len(s.response) > 0 {
		return s.response, nil
	} else {
		return nil, status.Error(codes.NotFound, "not found") //todo error?
	}
}

func (s *searchService) getFiles(req shared.SearchRequest) ([]string, error) {
	from := time.Date(req.From.Year(), req.From.Month(), req.From.Day(), 0, 0, 0, 0, req.From.Location())
	to := time.Date(req.To.Year(), req.To.Month(), req.To.Day(), 0, 0, 0, 0, req.To.Location())
	neededDir := make([]string, 0)
	for {
		if from.Before(to) {
			neededDir = append(neededDir, from.Format(dirLayout))
			from = from.Add(24 * time.Hour)
		} else {
			neededDir = append(neededDir, from.Format(dirLayout))
			break
		}
	}
	if len(neededDir) == 0 {
		return nil, status.Errorf(codes.NotFound, "not found directory from %s to %s", req.From.Format(dirLayout), req.To.Format(dirLayout))
	}

	response := make([]string, 0)
	baseDir := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	for _, dir := range neededDir {
		dir := baseDir + "/" + dir + "/" + req.ModuleName
		fileInfo, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		//todo check file name - address and time

		for _, file := range fileInfo {
			response = append(response, dir+"/"+file.Name())
		}
	}
	return response, nil
}

func (s *searchService) readFiles(req shared.SearchRequest, files []string) error {
	for _, filePath := range files {
		if err := s.unmarshalFile(req, filePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchService) unmarshalFile(req shared.SearchRequest, path string) error {
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(file)
	defer func() { _ = gzipReader.Close() }()
	if err != nil {
		return err
	}
	for {
		if s.count >= req.Limit {
			if s.offset == req.Offset {
				return nil
			} else {
				s.offset++
				s.count = 0
			}
		}
		response, err := entry.UnmarshalNext(gzipReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if checkEntry(req, response) {
			s.count++
			if s.offset == req.Offset {
				s.response = append(s.response, convertResponse(response))
			}
		}
	}
	return nil
}

func convertResponse(entries *entry.Entry) shared.SearchResponse {
	return shared.SearchResponse{
		ModuleName: entries.ModuleName,
		Host:       entries.Host,
		Event:      entries.Event,
		Level:      entries.Level,
		Time:       entries.Time,
		Request:    string(entries.Request),
		Response:   string(entries.Response),
		ErrorText:  entries.ErrorText,
	}
}

func checkEntry(req shared.SearchRequest, entries *entry.Entry) bool {
	if !checkEntryField(req.Level, entries.Level) {
		return false
	}
	if !checkEntryField(req.Host, entries.Host) {
		return false
	}
	if !checkEntryField(req.Event, entries.Event) {
		return false
	}
	return true
}

func checkEntryField(arrayOfExpected []string, field string) bool {
	for _, expected := range arrayOfExpected {
		if expected == field {
			return true
		}
	}
	return false
}
