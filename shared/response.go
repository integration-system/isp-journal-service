package shared

type SearchResponse struct {
	ModuleName string `json:",omitempty"`
	Host       string `json:",omitempty"`
	Event      string `json:",omitempty"`
	Level      string `json:",omitempty"`
	Time       string `json:",omitempty"`
	Request    string `json:",omitempty"`
	Response   string `json:",omitempty"`
	ErrorText  string `json:",omitempty"`
}
