package data

import (
	"encoding/json"
	"os"

	"github.com/ccss/internal/model"
)

// LoadStats reads and parses ~/.claude/stats-cache.json.
func LoadStats() (*model.StatsCache, error) {
	raw, err := os.ReadFile(StatsFile())
	if err != nil {
		return nil, err
	}
	var stats model.StatsCache
	if err := json.Unmarshal(raw, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
