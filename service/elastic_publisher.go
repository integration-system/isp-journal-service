package service

import (
	"os"

	"github.com/integration-system/isp-journal/entry"
	"github.com/integration-system/isp-journal/search"
	"github.com/integration-system/isp-lib/v2/config"
	log "github.com/integration-system/isp-log"
	"github.com/pkg/errors"
	"isp-journal-service/conf"
	"isp-journal-service/consts"
	"isp-journal-service/entity"
	"isp-journal-service/log_code"
	"isp-journal-service/model"
)

const limitStore = 1000

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

	fil, err := search.NewFilter(search.SearchRequest{})
	if err != nil {
		return err
	}
	logReader, err := search.NewLogReader(file, true, fil)
	if err != nil {
		return err
	}
	defer logReader.Close()

	for {
		logRecord, err := logReader.FilterNext()
		if err != nil {
			return err
		}
		if logRecord == nil {
			batch := s.store[:s.count]
			if len(batch) > 0 {
				err = s.insertBatch(batch)
				if err != nil {
					log.Error(log_code.ErrorElastic, err)
				}
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

	doc, err := json.Marshal(map[string]interface{}{
		"@timestamp": t,
		"moduleName": e.ModuleName,
		"host":       e.Host,
		"event":      e.Event,
		"level":      e.Level,
		"request":    string(e.Request),
		"response":   string(e.Response),
		"errorText":  e.ErrorText,
	})
	if err != nil {
		return entity.ElasticRecord{}, err
	}

	return entity.ElasticRecord{
		Index: consts.LogstashIndex,
		Doc:   doc,
	}, nil
}

func (s *elasticPublisher) publishRecord(record entity.ElasticRecord) {
	s.store[s.count] = record
	if s.count == limitStore-1 {
		err := s.insertBatch(s.store)
		if err != nil {
			log.Error(log_code.ErrorElastic, err)
		}
		s.count = 0
	} else {
		s.count++
	}
}

func (s *elasticPublisher) insertBatch(batch []entity.ElasticRecord) error {
	resp, err := model.Elastic.InsertBatch(batch)
	if err != nil {
		return err
	}
	if resp.Errors == true {
		return errors.New("insert error")
	}
	return nil
}
