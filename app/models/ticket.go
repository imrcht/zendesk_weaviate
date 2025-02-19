package models

type Ticket struct {
	ID       string `json:"id"`
	Metadata struct {
		Text     string   `json:"text"`
		LangCode string   `json:"lang_code"`
		Tags     []string `json:"tags"`
		Type     string   `json:"type"`
	} `json:"metadata"`
}
