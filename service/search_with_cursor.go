package service

import (
	"sync"
	"time"

	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-journal-service/conf"
)

type (
	cursorService struct {
		sync.Mutex
		cursorById map[string]*cursor
	}
	cursor struct {
		mx    sync.Mutex
		id    string
		s     *search.SyncSearchLog
		timer *time.Timer
	}
)

var CursorService = cursorService{
	cursorById: make(map[string]*cursor),
}

func (s *cursorService) Search(req *search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	if req.CursorId == "" {
		return s.newCursor(req)
	}

	CursorService.Lock()
	cursor, ok := CursorService.cursorById[req.CursorId]
	CursorService.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "cursor with id %s not found", req.CursorId)
	}
	cursor.timer.Stop()
	return cursor.nextBatch(req.BatchSize)
}

func (s *cursorService) newCursor(req *search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	cfg := config.GetRemote().(*conf.RemoteConfig)
	searchService, err := search.NewSyncSearchService(req.Request, cfg.BaseLogDirectory)
	if err != nil {
		return nil, err
	}

	cursor := &cursor{
		id:    uuid.NewV1().String(),
		s:     searchService,
		timer: time.NewTimer(time.Duration(cfg.CursorLifetime) * time.Second),
	}
	cursor.timer.Stop()
	CursorService.Lock()
	CursorService.cursorById[cursor.id] = cursor
	CursorService.Unlock()
	go cursor.deleteCursor()
	return cursor.nextBatch(req.BatchSize)
}

func (c *cursor) nextBatch(batchSize int) (*search.SearchWithCursorResponse, error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	defer c.reset()

	hasMoreEntries := true
	items := make([]search.SearchResponse, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		entry, hasMore, err := c.s.Next()
		if err != nil {
			return nil, err
		}
		if entry != nil {
			items = append(items, convertResponse(entry))
		}
		hasMoreEntries = hasMore
		if !hasMore {
			break
		}
	}
	return &search.SearchWithCursorResponse{
		CursorId: c.id,
		Items:    items,
		HasMore:  hasMoreEntries,
	}, nil
}

func (c *cursor) reset() {
	cursorLifeTime := config.GetRemote().(*conf.RemoteConfig).CursorLifetime
	c.timer.Reset(time.Duration(cursorLifeTime) * time.Second)
}

func (c *cursor) deleteCursor() {
	<-c.timer.C
	CursorService.Lock()
	delete(CursorService.cursorById, c.id)
	CursorService.Unlock()
}
