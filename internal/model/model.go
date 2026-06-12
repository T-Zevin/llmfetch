package model

type Model struct {
	Rank     int    `json:"rank"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	BestFor  string `json:"best_for"`
	Type     string `json:"type"`
	Score    int    `json:"score"`
	Runtime  string `json:"runtime"`
	OutTPS   int    `json:"out_tps"`
	InTPS    int    `json:"in_tps"`
	MemoryGB int    `json:"memory_gb"`
	Fit      string `json:"fit"`
	Context  string `json:"context"`
	License  string `json:"license"`
	Trend    int    `json:"trend"`
}

func FitRank(fit string) int {
	switch fit {
	case "Best":
		return 3
	case "Good":
		return 2
	case "Near":
		return 1
	default:
		return 0
	}
}
