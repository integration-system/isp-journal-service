package service

import (
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	"isp-journal-service/conf"
)

type searchService struct {
	counterLimit  int
	counterOffset int

	limit  int
	offset int

	response []search.SearchResponse
}

func NewSearchService() *searchService {
	return &searchService{
		counterLimit:  0,
		counterOffset: 0,
	}
}

func (s *searchService) Search(req *search.SearchRequest) ([]search.SearchResponse, error) {
	s.response = make([]search.SearchResponse, 0, req.Limit)
	s.limit = req.Limit
	s.offset = req.Offset
	baseDir := config.GetRemote().(*conf.RemoteConfig).BaseLogDirectory
	err := search.NewSearchLog(s.worker, baseDir).Search(*req)
	if err != nil {
		return nil, err
	}
	return s.response, nil
}

func (s *searchService) worker(entries *entry.Entry) (bool, error) {
	if s.counterLimit >= s.limit {
		return false, nil
	}
	if s.counterOffset == s.offset {
		s.counterLimit++
		s.response = append(s.response, convertResponse(entries))
	} else {
		s.counterOffset++
	}
	return true, nil
}

func convertResponse(entries *entry.Entry) search.SearchResponse {
	return search.SearchResponse{
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
