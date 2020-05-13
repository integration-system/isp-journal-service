package model

import (
	"context"

	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"github.com/olivere/elastic"
	uuid "github.com/satori/go.uuid"
	"isp-journal-service/conf"
	es "isp-journal-service/elastic"
	"isp-journal-service/entity"
	"isp-journal-service/log_code"
)

var Elastic = &elasticClient{
	cli: nil,
}

type elasticClient struct {
	cli *es.RxElasticClient
}

func (e *elasticClient) ReceiveConfiguration(setting conf.ElasticSetting) {
	if setting.Enable {
		if setting.Config == nil {
			log.Fatal(stdcodes.ModuleInvalidRemoteConfig, "await elastic configuration")
		}
		if e.cli == nil {
			e.defaultElasticClient()
		}
		e.cli.ReceiveConfiguration(setting)
	} else {
		if e.cli != nil {
			err := e.cli.Close()
			if err != nil {
				log.Error(log_code.ErrorElastic, err)
			}
			e.cli = nil
		}
	}
}

func (e *elasticClient) InsertBatch(records []entity.ElasticRecord) (*elastic.BulkResponse, error) {
	requestList := make([]elastic.BulkableRequest, len(records))
	for i, record := range records {
		requestList[i] = elastic.NewBulkIndexRequest().
			Index(record.Index).
			Doc(record.Doc).
			Type("_doc").
			Id(uuid.NewV1().String())
	}

	var (
		resp *elastic.BulkResponse
		err  error
	)
	err = e.cli.Visit(func(c *elastic.Client) error {
		resp, err = c.Bulk().Add(requestList...).Do(context.Background())
		return err
	})
	return resp, err
}

func (e *elasticClient) defaultElasticClient() {
	e.cli = es.NewRxElasticClient(
		es.WithInitializingHandler(e.initializingHandler),
		es.WithInitializingErrorHandler(e.errorHandler),
	)
}

func (e *elasticClient) initializingHandler(c *elastic.Client, config structure.ElasticConfiguration) {
	log.Infof(log_code.ErrorElastic, "elastic: successfully connected to %v", config.URL)
}

func (e *elasticClient) errorHandler(err *es.ErrorEvent) {
	log.WithMetadata(map[string]interface{}{
		"event": "initializing handler",
	}).Fatal(log_code.ErrorElastic, err)
}
