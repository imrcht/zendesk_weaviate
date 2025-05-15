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

type TicketCluster struct {
	TicketID           int    `json:"ticket_id"`
	Description        string `json:"description"`
	Composant          string `json:"composant"`
	DirectReason       string `json:"direct_reason"`
	DirectReasonOption string `json:"direct_reason_option"`
	Status             string `json:"status"`
}

type TicketClusterMap map[string][]TicketCluster
