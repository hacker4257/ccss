package model

import "time"

// ProjectGroup represents a Claude Code project directory with all its sessions.
type ProjectGroup struct {
	DirName      string         // e.g. "D--claudecode"
	OriginalPath string         // e.g. "D:\claudecode"
	Sessions     []SessionEntry // sorted by Modified desc
}

// SessionEntry corresponds to one entry in sessions-index.json.
type SessionEntry struct {
	SessionID    string    `json:"sessionId"`
	FullPath     string    `json:"fullPath"`
	FileMtime    int64     `json:"fileMtime"`
	FirstPrompt  string    `json:"firstPrompt"`
	Summary      string    `json:"summary"`
	MessageCount int       `json:"messageCount"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	GitBranch    string    `json:"gitBranch"`
	ProjectPath  string    `json:"projectPath"`
	IsSidechain  bool      `json:"isSidechain"`

	// Derived field (not in JSON)
	ProjectDirName string `json:"-"`
}

// SessionsIndex is the top-level structure of sessions-index.json.
type SessionsIndex struct {
	Version      int            `json:"version"`
	Entries      []SessionEntry `json:"entries"`
	OriginalPath string         `json:"originalPath"`
}
