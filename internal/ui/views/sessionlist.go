package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/model"
	"github.com/ccss/internal/theme"
)

// listItem represents either a project header or a session entry in the flat list.
type listItem struct {
	isHeader    bool
	projectName string
	session     *model.SessionEntry
}

// SessionListView displays all sessions grouped by project.
type SessionListView struct {
	items    []listItem // full list (headers + sessions)
	filtered []int      // indices into items that match filter
	cursor   int        // cursor position in filtered list
	offset   int        // scroll offset

	filterInput textinput.Model
	filtering   bool
	filterText  string

	totalSessions int
	width, height int
}

// NewSessionListView creates a new session list view.
func NewSessionListView() *SessionListView {
	ti := textinput.New()
	ti.Placeholder = "Filter sessions..."
	ti.CharLimit = 100

	return &SessionListView{
		filterInput: ti,
	}
}

// SetData populates the list with project groups.
func (v *SessionListView) SetData(groups []model.ProjectGroup) {
	v.items = nil
	v.totalSessions = 0
	for _, g := range groups {
		v.items = append(v.items, listItem{
			isHeader:    true,
			projectName: g.OriginalPath,
		})
		for i := range g.Sessions {
			v.items = append(v.items, listItem{
				session: &g.Sessions[i],
			})
			v.totalSessions++
		}
	}
	v.applyFilter()
}

func (v *SessionListView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

// SelectedSession returns the currently selected session, or nil.
func (v *SessionListView) SelectedSession() *model.SessionEntry {
	if v.cursor < 0 || v.cursor >= len(v.filtered) {
		return nil
	}
	item := v.items[v.filtered[v.cursor]]
	return item.session
}

// Update handles input for the session list.
func (v *SessionListView) Update(msg tea.Msg) tea.Cmd {
	if v.filtering {
		return v.updateFiltering(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.moveCursor(1)
		case "k", "up":
			v.moveCursor(-1)
		case "/":
			v.filtering = true
			v.filterInput.Focus()
			return textinput.Blink
		}
	}
	return nil
}

func (v *SessionListView) updateFiltering(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			v.filtering = false
			v.filterInput.Blur()
			v.filterText = ""
			v.filterInput.SetValue("")
			v.applyFilter()
			return nil
		case "enter":
			v.filtering = false
			v.filterInput.Blur()
			return nil
		}
	}

	var cmd tea.Cmd
	v.filterInput, cmd = v.filterInput.Update(msg)
	v.filterText = v.filterInput.Value()
	v.applyFilter()
	return cmd
}

func (v *SessionListView) applyFilter() {
	v.filtered = nil
	query := strings.ToLower(v.filterText)

	// Track which project headers should be shown
	var currentHeaderIdx int = -1
	headerHasMatch := false

	for i, item := range v.items {
		if item.isHeader {
			currentHeaderIdx = i
			headerHasMatch = false
			continue
		}
		if query == "" || v.matchesFilter(item, query) {
			if !headerHasMatch && currentHeaderIdx >= 0 {
				v.filtered = append(v.filtered, currentHeaderIdx)
				headerHasMatch = true
			}
			v.filtered = append(v.filtered, i)
		}
	}

	if v.cursor >= len(v.filtered) {
		v.cursor = len(v.filtered) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
	v.skipHeaders(1)
}

func (v *SessionListView) matchesFilter(item listItem, query string) bool {
	if item.session == nil {
		return false
	}
	return strings.Contains(strings.ToLower(item.session.FirstPrompt), query) ||
		strings.Contains(strings.ToLower(item.session.Summary), query)
}

func (v *SessionListView) moveCursor(delta int) {
	v.cursor += delta
	if v.cursor < 0 {
		v.cursor = 0
	}
	if v.cursor >= len(v.filtered) {
		v.cursor = len(v.filtered) - 1
	}
	v.skipHeaders(delta)
	v.ensureVisible()
}

func (v *SessionListView) skipHeaders(direction int) {
	if len(v.filtered) == 0 {
		return
	}
	for v.cursor >= 0 && v.cursor < len(v.filtered) {
		if !v.items[v.filtered[v.cursor]].isHeader {
			break
		}
		if direction >= 0 {
			v.cursor++
		} else {
			v.cursor--
		}
	}
	if v.cursor >= len(v.filtered) {
		v.cursor = len(v.filtered) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
}

func (v *SessionListView) ensureVisible() {
	visibleHeight := v.height - 4
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if v.cursor < v.offset {
		v.offset = v.cursor
	}
	if v.cursor >= v.offset+visibleHeight {
		v.offset = v.cursor - visibleHeight + 1
	}
}

// View renders the session list.
func (v *SessionListView) View() string {
	var b strings.Builder

	// Filter bar
	if v.filtering {
		b.WriteString("\n")
		b.WriteString(theme.FilterInputStyle.Render(v.filterInput.View()) + "\n\n")
	}

	if len(v.filtered) == 0 {
		b.WriteString(theme.EmptyStateStyle.Render("No sessions found.") + "\n")
		return b.String()
	}

	visibleHeight := v.height - 4
	if v.filtering {
		visibleHeight -= 3
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	end := v.offset + visibleHeight
	if end > len(v.filtered) {
		end = len(v.filtered)
	}

	for idx := v.offset; idx < end; idx++ {
		item := v.items[v.filtered[idx]]

		if item.isHeader {
			// Project group header with folder icon
			header := "  " + lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("▸ "+item.projectName)
			b.WriteString(header + "\n")
			continue
		}

		s := item.session
		isSelected := idx == v.cursor

		// Build the session line with proper structure
		timeStr := s.Modified.Local().Format("15:04")
		prompt := truncate(s.FirstPrompt, v.width-45)
		if prompt == "" {
			prompt = truncate(s.Summary, v.width-45)
		}
		countStr := fmt.Sprintf("%d msgs", s.MessageCount)
		branch := s.GitBranch

		// Cursor indicator
		cursor := "  "
		if isSelected {
			cursor = theme.SessionCursorStyle.Render("▸ ")
		}

		// Build components
		var line string
		if isSelected {
			line = cursor +
				lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(timeStr) + "  " +
				lipgloss.NewStyle().Foreground(theme.White).Bold(true).Render("\""+prompt+"\"") + "  " +
				lipgloss.NewStyle().Foreground(theme.TextDim).Render(countStr)
			if branch != "" && branch != "HEAD" {
				line += "  " + lipgloss.NewStyle().Foreground(theme.Green).Italic(true).Render(branch)
			}
			// Apply selected background to entire line
			line = lipgloss.NewStyle().Background(theme.BgHi).Width(v.width - 2).Render(line)
		} else {
			line = cursor +
				lipgloss.NewStyle().Foreground(theme.Cyan).Render(timeStr) + "  " +
				lipgloss.NewStyle().Foreground(theme.TextPri).Render("\""+prompt+"\"") + "  " +
				lipgloss.NewStyle().Foreground(theme.TextDim).Render(countStr)
			if branch != "" && branch != "HEAD" {
				line += "  " + lipgloss.NewStyle().Foreground(theme.Green).Italic(true).Render(branch)
			}
		}

		b.WriteString("  " + line + "\n")
	}

	// Bottom info bar
	b.WriteString("\n")
	sessionCount := v.totalSessions
	if v.filterText != "" {
		// Count only sessions (not headers) in filtered
		sessionCount = 0
		for _, idx := range v.filtered {
			if !v.items[idx].isHeader {
				sessionCount++
			}
		}
	}

	pct := 0
	if len(v.filtered) > 1 {
		pct = v.cursor * 100 / (len(v.filtered) - 1)
	}

	info := lipgloss.NewStyle().Foreground(theme.TextDim).Render(
		fmt.Sprintf("  %d sessions", sessionCount))
	if len(v.filtered) > visibleHeight {
		info += lipgloss.NewStyle().Foreground(theme.TextDim).Render(
			fmt.Sprintf("  ·  %d%%", pct))
	}
	if v.filterText != "" {
		info += lipgloss.NewStyle().Foreground(theme.Accent).Render(
			fmt.Sprintf("  ·  filter: %s", v.filterText))
	}
	b.WriteString(info)

	return b.String()
}

// StatusKeys returns the status bar text for this view.
func (v *SessionListView) StatusKeys() string {
	if v.filtering {
		return "enter:confirm  esc:cancel"
	}
	return "↑↓:navigate  enter:open  /:filter  ?:help  q:quit"
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
