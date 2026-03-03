package data

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/ccss/internal/model"
)

// LoadSessionIndex reads and parses a single sessions-index.json.
func LoadSessionIndex(projectDirName string) (*model.SessionsIndex, error) {
	path := SessionIndexPath(projectDirName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx model.SessionsIndex
	if err := json.Unmarshal(raw, &idx); err != nil {
		return nil, err
	}
	for i := range idx.Entries {
		idx.Entries[i].ProjectDirName = projectDirName
	}
	return &idx, nil
}

// LoadAllSessions loads session indices from all project directories
// and returns them as ProjectGroups sorted by most recent activity.
func LoadAllSessions() ([]model.ProjectGroup, error) {
	dirs, err := ListProjectDirs()
	if err != nil {
		return nil, err
	}
	var groups []model.ProjectGroup
	for _, dir := range dirs {
		idx, err := LoadSessionIndex(dir)
		if err != nil {
			continue
		}
		// Filter out sidechain sessions
		var sessions []model.SessionEntry
		for _, s := range idx.Entries {
			if !s.IsSidechain {
				sessions = append(sessions, s)
			}
		}
		if len(sessions) == 0 {
			continue
		}
		// Sort sessions by Modified descending
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].Modified.After(sessions[j].Modified)
		})
		groups = append(groups, model.ProjectGroup{
			DirName:      dir,
			OriginalPath: idx.OriginalPath,
			Sessions:     sessions,
		})
	}
	// Sort groups by most recent session
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Sessions[0].Modified.After(groups[j].Sessions[0].Modified)
	})
	return groups, nil
}
