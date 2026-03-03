package data

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"

	"github.com/ccss/internal/model"
)

// SearchResult represents a single search match.
type SearchResult struct {
	Session model.SessionEntry
	Snippet string
	IsDeep  bool // true if found via deep JSONL scan
}

// Search performs full-text search across all sessions.
// Phase 1: quick metadata search (firstPrompt, summary)
// Phase 2: deep JSONL content search (concurrent)
func Search(query string, allGroups []model.ProjectGroup) []SearchResult {
	query = strings.ToLower(query)
	if query == "" {
		return nil
	}

	var results []SearchResult

	// Phase 1: metadata search
	for _, g := range allGroups {
		for _, s := range g.Sessions {
			if strings.Contains(strings.ToLower(s.FirstPrompt), query) ||
				strings.Contains(strings.ToLower(s.Summary), query) {
				snippet := s.FirstPrompt
				if s.Summary != "" {
					snippet = s.Summary
				}
				results = append(results, SearchResult{
					Session: s,
					Snippet: truncateSnippet(snippet, 120),
				})
			}
		}
	}

	// Phase 2: deep content search
	var allSessions []model.SessionEntry
	for _, g := range allGroups {
		allSessions = append(allSessions, g.Sessions...)
	}

	// Track sessions already found in Phase 1
	found := make(map[string]bool)
	for _, r := range results {
		found[r.Session.SessionID] = true
	}

	sem := make(chan struct{}, 4)
	var mu sync.Mutex
	var wg sync.WaitGroup

	queryBytes := []byte(strings.ToLower(query))

	for _, s := range allSessions {
		if found[s.SessionID] {
			continue
		}
		wg.Add(1)
		go func(sess model.SessionEntry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			snippet := searchInFile(sess.FullPath, queryBytes)
			if snippet != "" {
				mu.Lock()
				results = append(results, SearchResult{
					Session: sess,
					Snippet: snippet,
					IsDeep:  true,
				})
				mu.Unlock()
			}
		}(s)
	}
	wg.Wait()

	return results
}

// searchInFile scans a JSONL file for the query in user/assistant message text.
func searchInFile(path string, queryBytes []byte) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Quick pre-filter: check if query appears in the raw line at all
		if !bytes.Contains(bytes.ToLower(line), queryBytes) {
			continue
		}
		// Only check user and assistant messages
		msgType := peekType(line)
		if msgType != "user" && msgType != "assistant" {
			continue
		}
		// Extract text content for snippet
		msg, err := parseMessage(line)
		if err != nil || msg == nil {
			continue
		}
		text := msg.DisplayText()
		lowerText := strings.ToLower(text)
		idx := strings.Index(lowerText, string(queryBytes))
		if idx >= 0 {
			return extractSnippet(text, idx, len(queryBytes), 120)
		}
	}
	return ""
}

// extractSnippet extracts a snippet around the match position.
func extractSnippet(text string, matchIdx, matchLen, maxLen int) string {
	runes := []rune(text)
	matchIdxRune := len([]rune(text[:matchIdx]))

	start := matchIdxRune - 40
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(runes) {
		end = len(runes)
	}

	snippet := string(runes[start:end])
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.ReplaceAll(snippet, "\r", "")

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet = snippet + "..."
	}
	return snippet
}

func truncateSnippet(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
