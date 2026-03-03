package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/data"
	"github.com/ccss/internal/model"
	"github.com/ccss/internal/theme"
)

// SearchView provides full-text search across sessions.
type SearchView struct {
	input   textinput.Model
	results []data.SearchResult
	cursor  int
	offset  int

	searching bool
	searched  bool
	groups    []model.ProjectGroup

	width, height int
}

func NewSearchView() *SearchView {
	ti := textinput.New()
	ti.Placeholder = "Search conversations..."
	ti.CharLimit = 200
	ti.Width = 60

	return &SearchView{
		input: ti,
	}
}

func (v *SearchView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.input.Width = w - 10
}

func (v *SearchView) SetData(groups []model.ProjectGroup) {
	v.groups = groups
}

func (v *SearchView) Focus() tea.Cmd {
	v.input.Focus()
	return textinput.Blink
}

func (v *SearchView) Blur() {
	v.input.Blur()
}

// InputFocused returns whether the search input is currently focused.
func (v *SearchView) InputFocused() bool {
	return v.input.Focused()
}

// SelectedSession returns the session of the selected search result, or nil.
func (v *SearchView) SelectedSession() *model.SessionEntry {
	if v.cursor < 0 || v.cursor >= len(v.results) {
		return nil
	}
	s := v.results[v.cursor].Session
	return &s
}

// SearchMsg is sent when search completes.
type SearchMsg struct {
	Results []data.SearchResult
}

func (v *SearchView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.input.Focused() {
			switch msg.String() {
			case "enter":
				query := v.input.Value()
				if query == "" {
					return nil
				}
				v.searching = true
				v.input.Blur()
				groups := v.groups
				return func() tea.Msg {
					results := data.Search(query, groups)
					return SearchMsg{Results: results}
				}
			case "down":
				if len(v.results) > 0 {
					v.input.Blur()
					v.cursor = 0
				}
				return nil
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if v.cursor > 0 {
					v.cursor--
					v.ensureVisible()
				} else {
					v.input.Focus()
					return textinput.Blink
				}
			case "down", "j":
				if v.cursor < len(v.results)-1 {
					v.cursor++
					v.ensureVisible()
				}
			case "/":
				v.input.Focus()
				return textinput.Blink
			}
		}

	case SearchMsg:
		v.searching = false
		v.searched = true
		v.results = msg.Results
		v.cursor = 0
		v.offset = 0
		return nil
	}

	if v.input.Focused() {
		var cmd tea.Cmd
		v.input, cmd = v.input.Update(msg)
		return cmd
	}
	return nil
}

func (v *SearchView) ensureVisible() {
	visibleHeight := v.height - 10
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	// Each result takes 3 lines
	maxVisible := visibleHeight / 3
	if maxVisible < 1 {
		maxVisible = 1
	}
	if v.cursor < v.offset {
		v.offset = v.cursor
	}
	if v.cursor >= v.offset+maxVisible {
		v.offset = v.cursor - maxVisible + 1
	}
}

func (v *SearchView) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// Search input
	b.WriteString(theme.SearchInputStyle.Render(v.input.View()) + "\n\n")

	if v.searching {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.Accent).PaddingLeft(4).Render("Searching...") + "\n")
		return b.String()
	}

	if !v.searched {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.TextDim).PaddingLeft(4).Render(
			"Type a query and press Enter to search across all sessions.") + "\n")
		return b.String()
	}

	if len(v.results) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(theme.TextDim).PaddingLeft(4).Render(
			"No results found.") + "\n")
		return b.String()
	}

	// Results count
	b.WriteString(lipgloss.NewStyle().Foreground(theme.TextSec).PaddingLeft(4).Render(
		fmt.Sprintf("%d results", len(v.results))) + "\n\n")

	visibleHeight := v.height - 10
	maxVisible := visibleHeight / 3
	if maxVisible < 1 {
		maxVisible = 1
	}

	end := v.offset + maxVisible
	if end > len(v.results) {
		end = len(v.results)
	}

	for i := v.offset; i < end; i++ {
		r := v.results[i]
		isSelected := i == v.cursor

		project := shortPath(r.Session.ProjectPath)
		prompt := truncate(r.Session.FirstPrompt, 40)
		date := r.Session.Modified.Local().Format("2006-01-02 15:04")

		// Cursor indicator
		cursor := "  "
		if isSelected {
			cursor = lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("> ")
		}

		var infoLine string
		if isSelected {
			infoLine = cursor +
				lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(project) +
				lipgloss.NewStyle().Foreground(theme.White).Bold(true).Render("  \""+prompt+"\"") +
				lipgloss.NewStyle().Foreground(theme.TextDim).Render("  "+date)
		} else {
			infoLine = cursor +
				lipgloss.NewStyle().Foreground(theme.Accent).Render(project) +
				lipgloss.NewStyle().Foreground(theme.TextPri).Render("  \""+prompt+"\"") +
				lipgloss.NewStyle().Foreground(theme.TextDim).Render("  "+date)
		}

		if isSelected {
			infoLine = lipgloss.NewStyle().Background(theme.BgHi).Width(v.width - 4).Render(infoLine)
		}
		b.WriteString("  " + infoLine + "\n")

		// Snippet
		snippet := truncate(r.Snippet, v.width-12)
		snippetLine := "      " + lipgloss.NewStyle().Foreground(theme.TextDim).Italic(true).Render(snippet)
		b.WriteString(snippetLine + "\n\n")
	}

	// Scroll indicator
	if len(v.results) > maxVisible {
		pct := 0
		if len(v.results) > 1 {
			pct = v.cursor * 100 / (len(v.results) - 1)
		}
		b.WriteString(lipgloss.NewStyle().Foreground(theme.TextDim).PaddingLeft(4).Render(
			fmt.Sprintf("%d/%d (%d%%)", v.cursor+1, len(v.results), pct)) + "\n")
	}

	return b.String()
}

func (v *SearchView) StatusKeys() string {
	if v.input.Focused() {
		return "enter:search  down:results  esc:back"
	}
	return "up/down:navigate  enter:open  /:search  esc:back"
}

// shortPath returns the last 2 components of a path for compact display.
func shortPath(p string) string {
	// Handle both / and \ separators
	sep := "/"
	if strings.Contains(p, "\\") {
		sep = "\\"
	}
	parts := strings.Split(p, sep)
	if len(parts) <= 2 {
		return p
	}
	return "..." + sep + strings.Join(parts[len(parts)-2:], sep)
}
