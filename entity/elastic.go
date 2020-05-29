package entity

import (
	"encoding/json"
)

type ElasticRecord struct {
	Index string
	Doc   json.RawMessage
}

type PolicyProgress struct {
	Indices map[string]Index
}

type Index struct {
	Phase string
}
