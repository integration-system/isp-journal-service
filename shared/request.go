package shared

import "time"

type SearchRequest struct {
	ModuleName string    //обязательно
	From       time.Time //2006-01-02T15:04:05.999-07:00"
	To         time.Time //2006-01-02T15:04:05.999-07:00"
	Host       []string  //опц
	Event      []string  //опц
	Level      []string  //опц
	Limit      int       //максимум 10000
	Offset     int       //отрицательное с конца
}
