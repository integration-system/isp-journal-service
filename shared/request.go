package shared

import "time"

type SearchRequest struct {
	ModuleName string `valid:"required~Required"`
	From       time.Time
	To         time.Time
	Host       []string
	Event      []string
	Level      []string
	Limit      int `valid:"required~Required,range(1|10000)"`
	Offset     int
}
