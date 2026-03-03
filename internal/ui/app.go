package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/data"
	"github.com/ccss/internal/model"
	"github.com/ccss/internal/theme"
	"github.com/ccss/internal/ui/views"
)

// ViewID identifies the current active view.
type ViewID int

const (
	ViewSessionList ViewID = iota
	ViewSessionDetail
	ViewSearch
	ViewStats
)

// App is the root Bubble Tea model.
type App struct {
	activeView ViewID
	showHelp   bool

	sessionList   *views.SessionListView
	sessionDetail *views.SessionDetailView
	searchView    *views.SearchView
	statsView     *views.StatsView
	helpView      *views.HelpView

	projectGroups []model.ProjectGroup
	statsCache    *model.StatsCache

	statusMsg string
	width     int
	height    int
}

// New creates a new App model.
func New() *App {
	return &App{
		activeView:    ViewSessionList,
		sessionList:   views.NewSessionListView(),
		sessionDetail: views.NewSessionDetailView(),
		searchView:    views.NewSearchView(),
		statsView:     views.NewStatsView(),
		helpView:      views.NewHelpView(),
	}
}

// --- Messages ---

type sessionsLoadedMsg struct {
	groups []model.ProjectGroup
}

type statsLoadedMsg struct {
	stats *model.StatsCache
}

type messagesLoadedMsg struct {
	session  model.SessionEntry
	messages []*model.Message
}

type exportDoneMsg struct {
	path string
	err  error
}

// --- Init ---

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		loadSessionsCmd(),
		loadStatsCmd(),
	)
}

func loadSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		groups, err := data.LoadAllSessions()
		if err != nil {
			return sessionsLoadedMsg{groups: nil}
		}
		return sessionsLoadedMsg{groups: groups}
	}
}

func loadStatsCmd() tea.Cmd {
	return func() tea.Msg {
		stats, err := data.LoadStats()
		if err != nil {
			return statsLoadedMsg{stats: nil}
		}
		return statsLoadedMsg{stats: stats}
	}
}

func loadMessagesCmd(session model.SessionEntry) tea.Cmd {
	return func() tea.Msg {
		messages, err := data.LoadSessionMessages(session.FullPath)
		if err != nil {
			return messagesLoadedMsg{session: session, messages: nil}
		}
		return messagesLoadedMsg{session: session, messages: messages}
	}
}

func exportCmd(session model.SessionEntry, messages []*model.Message) tea.Cmd {
	return func() tea.Msg {
		content := data.ExportSessionMarkdown(session, messages)
		shortID := session.SessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		path := fmt.Sprintf("ccss-export-%s.md", shortID)
		err := data.WriteExportFile(path, content)
		return exportDoneMsg{path: path, err: err}
	}
}

// --- Update ---

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentHeight := a.height - 4 // header(2) + status(1) + separator(1)
		if contentHeight < 1 {
			contentHeight = 1
		}
		a.sessionList.SetSize(a.width, contentHeight)
		a.sessionDetail.SetSize(a.width, contentHeight)
		a.searchView.SetSize(a.width, contentHeight)
		a.statsView.SetSize(a.width, contentHeight)
		a.helpView.SetSize(a.width, contentHeight)
		return a, nil

	case sessionsLoadedMsg:
		a.projectGroups = msg.groups
		a.sessionList.SetData(msg.groups)
		a.searchView.SetData(msg.groups)
		return a, nil

	case statsLoadedMsg:
		a.statsCache = msg.stats
		if msg.stats != nil {
			a.statsView.SetData(msg.stats)
		}
		return a, nil

	case messagesLoadedMsg:
		a.sessionDetail.SetSession(msg.session, msg.messages)
		a.activeView = ViewSessionDetail
		a.statusMsg = ""
		return a, nil

	case exportDoneMsg:
		if msg.err != nil {
			a.statusMsg = "Export failed: " + msg.err.Error()
		} else {
			a.statusMsg = "Exported to " + msg.path
		}
		return a, nil

	case tea.KeyMsg:
		// Help toggle is global
		if msg.String() == "?" && a.activeView != ViewSearch {
			a.showHelp = !a.showHelp
			return a, nil
		}
		if a.showHelp {
			if msg.String() == "esc" || msg.String() == "?" {
				a.showHelp = false
			}
			return a, nil
		}

		// Global quit
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		// Global tab switching (only when not in detail view or search input)
		if a.activeView != ViewSessionDetail {
			switch msg.String() {
			case "1":
				a.activeView = ViewSessionList
				a.searchView.Blur()
				a.statusMsg = ""
				return a, nil
			case "2":
				a.activeView = ViewSearch
				a.statusMsg = ""
				return a, a.searchView.Focus()
			case "3":
				a.activeView = ViewStats
				a.searchView.Blur()
				a.statusMsg = ""
				return a, nil
			}
		}

		// View-specific handling
		return a.updateActiveView(msg)
	}

	// Pass non-key messages to active view
	return a.updateActiveViewMsg(msg)
}

func (a *App) updateActiveView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.activeView {
	case ViewSessionList:
		switch msg.String() {
		case "q":
			return a, tea.Quit
		case "enter":
			s := a.sessionList.SelectedSession()
			if s != nil {
				a.statusMsg = "Loading session..."
				return a, loadMessagesCmd(*s)
			}
		}
		cmd := a.sessionList.Update(msg)
		return a, cmd

	case ViewSessionDetail:
		switch msg.String() {
		case "esc":
			a.activeView = ViewSessionList
			a.statusMsg = ""
			return a, nil
		case "e":
			session := a.sessionDetail.Session()
			messages := a.sessionDetail.Messages()
			if len(messages) > 0 {
				return a, exportCmd(session, messages)
			}
			return a, nil
		case "1":
			a.activeView = ViewSessionList
			a.statusMsg = ""
			return a, nil
		case "2":
			a.activeView = ViewSearch
			a.statusMsg = ""
			return a, a.searchView.Focus()
		case "3":
			a.activeView = ViewStats
			a.statusMsg = ""
			return a, nil
		}
		cmd := a.sessionDetail.Update(msg)
		return a, cmd

	case ViewSearch:
		switch msg.String() {
		case "esc":
			a.activeView = ViewSessionList
			a.searchView.Blur()
			a.statusMsg = ""
			return a, nil
		}
		// Let search view handle enter (for both search execution and result selection)
		cmd := a.searchView.Update(msg)
		// After update, check if user selected a result (input not focused + enter)
		if msg.String() == "enter" {
			s := a.searchView.SelectedSession()
			if s != nil && !a.searchView.InputFocused() {
				a.statusMsg = "Loading session..."
				return a, loadMessagesCmd(*s)
			}
		}
		return a, cmd

	case ViewStats:
		switch msg.String() {
		case "esc":
			a.activeView = ViewSessionList
			a.statusMsg = ""
			return a, nil
		}
		cmd := a.statsView.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) updateActiveViewMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.activeView {
	case ViewSearch:
		cmd := a.searchView.Update(msg)
		return a, cmd
	case ViewSessionDetail:
		cmd := a.sessionDetail.Update(msg)
		return a, cmd
	case ViewStats:
		cmd := a.statsView.Update(msg)
		return a, cmd
	}
	return a, nil
}

// --- View ---

func (a *App) View() string {
	header := a.renderHeader()
	var content string
	if a.showHelp {
		content = a.helpView.View()
	} else {
		content = a.renderActiveView()
	}
	statusBar := a.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

func (a *App) renderHeader() string {
	// Title
	title := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Render("CCSS")

	// Tabs
	tabs := []struct {
		label string
		view  ViewID
	}{
		{"Sessions", ViewSessionList},
		{"Search", ViewSearch},
		{"Stats", ViewStats},
	}

	var tabParts []string
	for i, tab := range tabs {
		label := fmt.Sprintf("%d %s", i+1, tab.label)
		isActive := a.activeView == tab.view ||
			(a.activeView == ViewSessionDetail && tab.view == ViewSessionList)
		if isActive {
			tabParts = append(tabParts, theme.ActiveTabStyle.Render(label))
		} else {
			tabParts = append(tabParts, theme.InactiveTabStyle.Render(label))
		}
	}

	// Breadcrumb for detail view
	breadcrumb := ""
	if a.activeView == ViewSessionDetail {
		s := a.sessionDetail.Session()
		breadcrumb = lipgloss.NewStyle().Foreground(theme.TextDim).Render(
			" > " + truncateStr(s.FirstPrompt, 40))
	}

	headerLine := " " + title + "  " +
		strings.Join(tabParts, " ") + breadcrumb

	// Separator line
	sep := lipgloss.NewStyle().Foreground(theme.Border).Render(
		safeRepeat("─", a.width))

	return headerLine + "\n" + sep
}

func (a *App) renderActiveView() string {
	var content string
	switch a.activeView {
	case ViewSessionList:
		content = a.sessionList.View()
	case ViewSessionDetail:
		content = a.sessionDetail.View()
	case ViewSearch:
		content = a.searchView.View()
	case ViewStats:
		content = a.statsView.View()
	}

	// Ensure content fills the available height
	contentHeight := a.height - 4
	if contentHeight > 0 {
		lines := strings.Count(content, "\n") + 1
		if lines < contentHeight {
			content += strings.Repeat("\n", contentHeight-lines)
		}
	}
	return content
}

func (a *App) renderStatusBar() string {
	var keys string
	if a.showHelp {
		keys = a.helpView.StatusKeys()
	} else {
		switch a.activeView {
		case ViewSessionList:
			keys = a.sessionList.StatusKeys()
		case ViewSessionDetail:
			keys = a.sessionDetail.StatusKeys()
		case ViewSearch:
			keys = a.searchView.StatusKeys()
		case ViewStats:
			keys = a.statsView.StatusKeys()
		}
	}

	barContent := " " + lipgloss.NewStyle().Foreground(theme.TextDim).Render(keys)
	if a.statusMsg != "" {
		barContent += "  " + lipgloss.NewStyle().Foreground(theme.Yellow).Render(a.statusMsg)
	}

	return lipgloss.NewStyle().Foreground(theme.TextDim).Width(a.width).Render(barContent)
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func safeRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}
