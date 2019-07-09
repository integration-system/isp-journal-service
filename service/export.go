package service

import (
	"encoding/csv"
	"github.com/integration-system/isp-journal/entry"
	"isp-journal-service/model"
	"isp-journal-service/shared"
	"isp-journal-service/utils"
)

var awaitingExport = []string{"ModuleName", "Host", "Event", "Level", "Time", "Request", "Response", "ErrorText"}

type exportService struct {
	shared.SearchRequest

	counterLimit  int
	counterOffset int

	writer *csv.Writer
}

func NewImportService(req shared.SearchRequest) *exportService {
	return &exportService{
		SearchRequest: req,
		counterLimit:  0,
		counterOffset: 0,
	}
}

func (s *exportService) Export(filepath string) error {
	return utils.UseCsvWriter(filepath, ';', s.exportLog)
}

func (s *exportService) exportLog(writer *csv.Writer) error {
	s.writer = writer
	if err := s.writer.Write(awaitingExport); err != nil {
		return err
	}
	if s.Limit == 0 {
		return model.NewSearchLog(s.workerWithoutLimit).Search(s.SearchRequest)
	} else {
		return model.NewSearchLog(s.workerWithLimit).Search(s.SearchRequest)
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
