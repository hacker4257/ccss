package data

import (
	"os"
	"path/filepath"
)

// ClaudeDir returns the absolute path to ~/.claude/.
func ClaudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// ProjectsDir returns ~/.claude/projects/.
func ProjectsDir() string {
	return filepath.Join(ClaudeDir(), "projects")
}

// StatsFile returns ~/.claude/stats-cache.json.
func StatsFile() string {
	return filepath.Join(ClaudeDir(), "stats-cache.json")
}

// HistoryFile returns ~/.claude/history.jsonl.
func HistoryFile() string {
	return filepath.Join(ClaudeDir(), "history.jsonl")
}

// ListProjectDirs returns all immediate subdirectories of ~/.claude/projects/.
func ListProjectDirs() ([]string, error) {
	entries, err := os.ReadDir(ProjectsDir())
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs, nil
}

// SessionIndexPath returns the sessions-index.json path for a project dir.
func SessionIndexPath(projectDirName string) string {
	return filepath.Join(ProjectsDir(), projectDirName, "sessions-index.json")
}
