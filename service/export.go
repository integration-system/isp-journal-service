package service

import (
	"encoding/csv"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/resources"
	"isp-journal-service/conf"
)

var awaitingExport = []string{"ModuleName", "Host", "Event", "Level", "Time", "Request", "Response", "ErrorText"}

type exportService struct {
	search.SearchRequest

	counterLimit  int
	counterOffset int

	writer *csv.Writer
}

func NewImportService(req search.SearchRequest) *exportService {
	return &exportService{
		SearchRequest: req,
		counterLimit:  0,
		counterOffset: 0,
	}
}

func (s *exportService) Export(filepath string) error {
	return resources.CompressedCsvWriter(filepath, ';', s.exportLog)
}

func (s *exportService) exportLog(writer *csv.Writer) error {
	s.writer = writer
	if err := s.writer.Write(awaitingExport); err != nil {
		return err
	}
	baseDir := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	if s.Limit == 0 {
		return search.NewSearchLog(s.workerWithoutLimit, baseDir).Search(s.SearchRequest)
	} else {
		return search.NewSearchLog(s.workerWithLimit, baseDir).Search(s.SearchRequest)
	}
}

func (s *exportService) workerWithLimit(entries *entry.Entry) (bool, error) {
	if s.counterLimit >= s.Limit {
		return false, nil
	}
	if s.counterOffset == s.Offset {
		s.counterLimit++
		if err := s.writeHandler(entries); err != nil {
			return false, err
		}
	} else {
		s.counterOffset++
	}
	return true, nil
}

func (s *exportService) workerWithoutLimit(entries *entry.Entry) (bool, error) {
	if err := s.writeHandler(entries); err != nil {
		return false, err
	}
	return true, nil
}

func (s *exportService) writeHandler(entries *entry.Entry) error {
	return s.writer.Write(s.convertEntries(entries))
}

func (s *exportService) convertEntries(e *entry.Entry) []string {
	return []string{e.ModuleName, e.Host, e.Event, e.Level, e.Time, string(e.Request), string(e.Response), e.ErrorText}
}
