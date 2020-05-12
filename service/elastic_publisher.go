package service

import (
	"encoding/json"
	"fmt"
	"os"

	io2 "github.com/integration-system/isp-io"
	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	log "github.com/integration-system/isp-log"
	"isp-journal-service/conf"
	"isp-journal-service/entity"
	"isp-journal-service/log_code"
	"isp-journal-service/model"
)

const (
	limitStore     = 1000
	logstashLayout = "2006-01-02"
)

func NewElasticPublisher() *elasticPublisher {
	return &elasticPublisher{
		store: make([]entity.ElasticRecord, limitStore),
		count: 0,
	}
}

type elasticPublisher struct {
	count int
	store []entity.ElasticRecord
}

func (s *elasticPublisher) Publish(dir string) error {
	if !config.GetRemote().(*conf.RemoteConfig).ElasticSetting.Enable {
		return nil
	}

	file, err := os.Open(dir)
	if err != nil {
		return err
	}

	readPipe := io2.NewReadPipe(file)
	logReader, err := search.NewLogReader(readPipe, true, search.Filter{})
	if err != nil {
		return err
	}

	for {
		logRecord, err := logReader.FilterNext()
		if err != nil {
			return err
		}
		if logRecord == nil {
			_, err := model.Elastic.InsertBatch(s.store[:s.count])
			if err != nil {
				log.Error(log_code.ErrorElastic, err)
			}
			break
		}

		elasticRecord, err := s.getElasticRecord(logRecord)
		if err != nil {
			return err
		}

		s.publishRecord(elasticRecord)
	}
	return nil
}

func (s *elasticPublisher) getElasticRecord(e *entry.Entry) (entity.ElasticRecord, error) {
	t, err := entry.ParserTime(e.Time)
	if err != nil {
		return entity.ElasticRecord{}, err
	}
	index := fmt.Sprintf("logstash-%s", t.Format(logstashLayout))

	doc, err := json.Marshal(map[string]interface{}{
		"@timestamp": t,
		"moduleName": e.ModuleName,
		"host":       e.Host,
		"event":      e.Event,
		"level":      e.Level,
		"time":       e.Request,
		"request":    s.formatBytes(e.Request),
		"response":   s.formatBytes(e.Response),
		"errorText":  e.ErrorText,
	})
	if err != nil {
		return entity.ElasticRecord{}, err
	}

	return entity.ElasticRecord{
		Index: index,
		Doc:   doc,
	}, nil
}

func (s *elasticPublisher) formatBytes(b []byte) interface{} {
	bytesLength := len(b)
	if bytesLength > 0 &&
		((b[0] == '{' && b[bytesLength-1] == '}') ||
			(b[0] == '[' && b[bytesLength-1] == ']')) {
		return json.RawMessage(b)
	}
	return string(b)
}

func (s *elasticPublisher) publishRecord(record entity.ElasticRecord) {
	s.store[s.count] = record
	if s.count == limitStore {
		_, err := model.Elastic.InsertBatch(s.store)
		if err != nil {
			log.Error(log_code.ErrorElastic, err)
		}
		s.count = 0
	}
	s.count++
}
