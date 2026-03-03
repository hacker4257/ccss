package model

// StatsCache is the top-level structure of stats-cache.json.
type StatsCache struct {
	Version          int                   `json:"version"`
	LastComputedDate string                `json:"lastComputedDate"`
	DailyActivity    []DailyActivity       `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens    `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsage `json:"modelUsage"`
	TotalSessions    int                   `json:"totalSessions"`
	TotalMessages    int                   `json:"totalMessages"`
	LongestSession   LongestSession        `json:"longestSession"`
	FirstSessionDate string                `json:"firstSessionDate"`
	HourCounts       map[string]int        `json:"hourCounts"`
}

// DailyActivity represents one day's activity metrics.
type DailyActivity struct {
	Date          string `json:"date"`
	MessageCount  int    `json:"messageCount"`
	SessionCount  int    `json:"sessionCount"`
	ToolCallCount int    `json:"toolCallCount"`
}

// DailyModelTokens represents token usage by model for one day.
type DailyModelTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

// ModelUsage represents cumulative usage for a single model.
type ModelUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
	WebSearchRequests        int `json:"webSearchRequests"`
}

// LongestSession tracks the single longest session.
type LongestSession struct {
	SessionID    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
	Timestamp    string `json:"timestamp"`
}
