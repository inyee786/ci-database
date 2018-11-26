package handler

// Pipeline struct
type Pipeline []struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// dashboardSummary contains the details of a gitlab pipelines
type dashboardSummary struct {
	ID     int    `json:"id"`
	Sha    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

type dashboardaws struct {
	Dashboardaws []dashboardSummary
}

type dashboard struct {
	Dashboard []dashboardSummary
}
