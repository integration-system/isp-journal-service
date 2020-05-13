package elastic

import (
	"errors"
	"fmt"
	"sync"

	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"isp-journal-service/conf"
	"isp-journal-service/log_code"
)

var ErrNotConnected = errors.New("elastic: not connected")

type ErrorEvent struct {
	action string
	err    error
	config structure.ElasticConfiguration
}

func (er ErrorEvent) Error() string {
	return fmt.Sprintf("rxElasticClient: %s: %v, config: %v", er.action, er.err, er.config)
}

type errorHandler func(err *ErrorEvent)

type initHandler func(c *elastic.Client, config structure.ElasticConfiguration)

type visitor func(c *elastic.Client) error

type RxElasticClient struct {
	cli *elastic.Client

	lastConf structure.ElasticConfiguration
	lock     sync.RWMutex
	active   bool

	initHandler initHandler
	eh          errorHandler
}

func (rc *RxElasticClient) ReceiveConfiguration(setting conf.ElasticSetting) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if !rc.active {
		return
	}

	if !cmp.Equal(rc.lastConf, setting.Config) {
		ok := true

		client, err := newElasticClient(setting.Config)
		if err != nil {
			if rc.eh != nil {
				rc.eh(&ErrorEvent{"connect", err, *setting.Config})
			}
			ok = false
		}

		if ok && rc.cli != nil {
			rc.cli.Stop()
			rc.cli = nil
		}

		if ok {
			rc.cli = client
			rc.lastConf = *setting.Config
			if rc.initHandler != nil {
				rc.initHandler(rc.cli, *setting.Config)
			}

			err = policy.CreateLogstashPolicy(setting.Config.URL, setting.Policy)
			if err != nil {
				log.Fatal(log_code.ErrorElastic, err)
			}
		}
	}
}

func (rc *RxElasticClient) Close() error {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	rc.active = false
	if rc.cli != nil {
		c := rc.cli
		rc.cli = nil
		c.Stop()
	}
	return nil
}

func (rc *RxElasticClient) Visit(v visitor) error {
	rc.lock.RLock()
	defer rc.lock.RUnlock()

	if rc.cli == nil {
		return ErrNotConnected
	}
	return v(rc.cli)
}

func NewRxElasticClient(opts ...Option) *RxElasticClient {
	rdc := &RxElasticClient{active: true}

	for _, o := range opts {
		o(rdc)
	}
	return rdc
}

func newElasticClient(cfg *structure.ElasticConfiguration) (*elastic.Client, error) {
	elasticConfig := config.Config{}
	err := cfg.ConvertTo(&elasticConfig)
	if err != nil {
		return nil, err
	}
	return elastic.NewClientFromConfig(&elasticConfig)
}
