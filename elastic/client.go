package elastic

import (
	"errors"
	"fmt"
	"sync"

	"github.com/integration-system/go-cmp/cmp"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
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
	c *elastic.Client

	lastConf structure.ElasticConfiguration
	lock     sync.RWMutex
	active   bool

	initHandler initHandler
	eh          errorHandler
}

func (rc *RxElasticClient) ReceiveConfiguration(config *structure.ElasticConfiguration) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if !rc.active {
		return
	}

	if !cmp.Equal(rc.lastConf, config) {
		ok := true

		var c *elastic.Client

		if client, err := newElasticClient(config); err != nil {
			if rc.eh != nil {
				rc.eh(&ErrorEvent{"connect", err, *config})
			}
			ok = false
		} else {
			c = client
		}

		if ok && rc.c != nil {
			rc.c.Stop()
			rc.c = nil
		}

		if ok {
			rc.c = c
			rc.lastConf = *config
			if rc.initHandler != nil {
				rc.initHandler(rc.c, *config)
			}
		}
	}
}

func (rc *RxElasticClient) Close() error {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	rc.active = false
	if rc.c != nil {
		c := rc.c
		rc.c = nil
		c.Stop()
	}
	return nil
}

func (rc *RxElasticClient) Visit(v visitor) error {
	rc.lock.RLock()
	defer rc.lock.RUnlock()

	if rc.c == nil {
		return ErrNotConnected
	}

	return v(rc.c)
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
	if err := cfg.ConvertTo(&elasticConfig); err != nil {
		return nil, err
	} else {
		return elastic.NewClientFromConfig(&elasticConfig)
	}
}
