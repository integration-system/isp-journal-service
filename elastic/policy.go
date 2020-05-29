package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/integration-system/isp-lib/v2/http"
	"isp-journal-service/conf"
	"isp-journal-service/consts"
	"isp-journal-service/entity"
)

const (
	newPolicyRequest = `
{
  "policy": {                       
    "phases": {
      "hot": {                      
        "actions": {
          "rollover": {             
            "max_size": "%s",
            "max_age": "%s"
          }
        }
      },
      "delete": {
        "min_age": "%s",           
        "actions": {
          "delete": {}              
        }
      }
    }
  }
}`
	indexTemplateRequest = `
{
  "index_patterns": ["%s-*"],                 
  "settings": {
    "number_of_shards": 4,
    "number_of_replicas": 0,
    "index.lifecycle.name": "%s_policy",      
    "index.lifecycle.rollover_alias": "%s"    
  }
}`
	writeFirstIndexRequest = `
{
  "aliases": {
    "%s": {
      "is_write_index": true
    }
  }
}`
)

var policy = &policySetting{
	cli: http.NewJsonRestClient(),
	headers: map[string]string{
		"Content-Type": "application/json",
	},
}

type policySetting struct {
	cli     http.RestClient
	headers map[string]string
}

func (s policySetting) CreateLogstashPolicy(uri string, setting conf.PolicySetting) error {
	err := s.newPolicy(uri, setting.RolloverSize, setting.RolloverAge, setting.DeleteAge)
	if err != nil {
		return err
	}

	exist, err := s.checkExistIndex(uri)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	err = s.indexTemplate(uri)
	if err != nil {
		return err
	}

	err = s.writeFirstIndex(uri)
	if err != nil {
		return err
	}
	return nil
}

func (s policySetting) newPolicy(uri, rolloverSize, rolloverAge, deleteAge string) error {
	return s.cli.Invoke(
		"PUT",
		fmt.Sprintf("%s/_ilm/policy/%s_policy", uri, consts.LogstashIndex),
		s.headers,
		json.RawMessage(fmt.Sprintf(newPolicyRequest, rolloverSize, rolloverAge, deleteAge)),
		nil,
	)
}

func (s policySetting) indexTemplate(uri string) error {
	return s.cli.Invoke(
		"PUT",
		fmt.Sprintf("%s/_template/%s_template", uri, consts.LogstashIndex),
		s.headers,
		json.RawMessage(fmt.Sprintf(indexTemplateRequest, consts.LogstashIndex, consts.LogstashIndex, consts.LogstashIndex)),
		nil,
	)
}

func (s policySetting) writeFirstIndex(uri string) error {
	return s.cli.Invoke(
		"PUT",
		fmt.Sprintf("%s/%s-000001", uri, consts.LogstashIndex),
		s.headers,
		json.RawMessage(fmt.Sprintf(writeFirstIndexRequest, consts.LogstashIndex)),
		nil,
	)
}

func (s policySetting) checkExistIndex(uri string) (bool, error) {
	response := new(entity.PolicyProgress)
	err := s.cli.Invoke(
		"GET",
		fmt.Sprintf("%s/%s-*/_ilm/explain", uri, consts.LogstashIndex),
		s.headers,
		nil,
		response,
	)
	if err != nil {
		return false, err
	}
	for _, index := range response.Indices {
		if index.Phase == "hot" {
			return true, nil
		}
	}
	return false, nil
}
