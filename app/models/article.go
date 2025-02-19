package models

// Section represents a section in the Zendesk knowledge base
type Section struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
}

// Category represents a category in the Zendesk knowledge base
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Article represents an article in the Zendesk knowledge base
type Article struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	HTMLURL string   `json:"html_url"`
	Locale  string   `json:"locale"`
	Locales []string `json:"locales"`
}
