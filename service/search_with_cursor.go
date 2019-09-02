package service

import (
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/config"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-journal-service/conf"
	"sync"
	"time"
)

type (
	cursorService struct {
		batchSize int
		counter   int
		response  *search.SearchWithCursorResponse
	}
	cursorStore struct {
		sync.Mutex
		cursorById map[string]*cursor
	}
	cursor struct {
		s     *search.SyncSearchLog
		timer *time.Timer
	}
)

var CursorStore = cursorStore{
	cursorById: make(map[string]*cursor),
}

func NewSearchWithCursor() *cursorService {
	return &cursorService{
		counter:   0,
		batchSize: 0,
		response:  &search.SearchWithCursorResponse{Items: make([]search.SearchResponse, 0)},
	}
}

func (s *cursorService) Search(req search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	if req.CursorId == "" {
		return s.newCursor(req)
	} else {
		CursorStore.Lock()
		cursor, ok := CursorStore.cursorById[req.CursorId]
		CursorStore.Unlock()
		if !ok {
			return nil, status.Errorf(codes.NotFound, "cursor with id %s not found", req.CursorId)
		} else {
			s.batchSize = req.BatchSize
			cursor.timer.Stop()
			s.response.CursorId = req.CursorId
			return s.worker(cursor)
		}
	}
}

func (s *cursorService) newCursor(req search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	s.batchSize = req.BatchSize
	cfg := config.GetRemote().(*conf.RemoteConfig)
	if searchService, err := search.NewSyncSearchService(req.Request, cfg.BaseLogDirectory); err != nil {
		return nil, err
	} else {
		s.response.CursorId = uuid.NewV1().String()
		newCursor := &cursor{
			s:     searchService,
			timer: time.NewTimer(time.Duration(cfg.CursorLifetime) * time.Second),
		}
		newCursor.timer.Stop()
		CursorStore.Lock()
		CursorStore.cursorById[s.response.CursorId] = newCursor
		CursorStore.Unlock()
		go newCursor.deleteCursor(s.response.CursorId)
		return s.worker(newCursor)
	}
}

func (s *cursorService) worker(cursor *cursor) (*search.SearchWithCursorResponse, error) {
	defer cursor.reset()

	s.counter = 0
	for {
		if s.counter < s.batchSize {
			extractedEntry, hasMore, err := cursor.s.Next()
			if err != nil {
				return nil, err
			}

			if extractedEntry != nil {
				s.response.Items = append(s.response.Items, convertResponse(extractedEntry))
				s.counter++
			}

			s.response.HasMore = hasMore
			if !hasMore {
				return s.response, nil
			}
		} else {
			return s.response, nil
		}
	}
}

func (c *cursor) reset() {
	cursorLifeTime := config.GetRemote().(*conf.RemoteConfig).CursorLifetime
	c.timer.Reset(time.Duration(cursorLifeTime) * time.Second)
}

func (c *cursor) deleteCursor(cursorId string) {
	<-c.timer.C
	CursorStore.Lock()
	delete(CursorStore.cursorById, cursorId)
	CursorStore.Unlock()
}
