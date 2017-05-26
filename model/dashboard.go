package model

type Target struct {
	Ns          string `json:"ns"`
	Measurement string `json:"measurement"`
	Fn          string `json:"function"`
	Where       string `json:"where"`
	// GroupBy     string `json:"groupby"`
}

type Panel struct {
	Title     string   `json:"title"`
	GraphType string   `json:"type"`
	Targets   []Target `json:"targets"`
	// Grid string
	// Span int
	// Fill string
}

type Dashboard struct {
	Title  string  `json:"title"`
	Panels []Panel `json:"panels"`
}

type DashboardData []Dashboard
