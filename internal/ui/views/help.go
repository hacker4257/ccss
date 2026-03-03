package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/theme"
)

// HelpView shows a keybinding reference overlay.
type HelpView struct {
	width, height int
}

func NewHelpView() *HelpView {
	return &HelpView{}
}

func (v *HelpView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

func (v *HelpView) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (v *HelpView) View() string {
	var b strings.Builder

	b.WriteString("\n")

	titleStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	b.WriteString("  " + titleStyle.Render("CCSS - Claude Code Session Manager") + "\n\n")

	section := func(title string, keys [][2]string) {
		b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(title) + "\n")
		b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Border).Render(safeRepeat("-", 36)) + "\n")
		for _, kv := range keys {
			key := lipgloss.NewStyle().Foreground(theme.Yellow).Bold(true).Width(16).PaddingLeft(2).Render(kv[0])
			desc := lipgloss.NewStyle().Foreground(theme.TextSec).Render(kv[1])
			b.WriteString("  " + key + desc + "\n")
		}
		b.WriteString("\n")
	}

	section("Global", [][2]string{
		{"1", "Sessions tab"},
		{"2", "Search tab"},
		{"3", "Stats tab"},
		{"?", "Toggle help"},
		{"ctrl+c", "Quit"},
	})

	section("Session List", [][2]string{
		{"up/down j/k", "Navigate"},
		{"enter", "Open session"},
		{"/", "Filter sessions"},
		{"esc", "Clear filter / back"},
		{"q", "Quit"},
	})

	section("Session Detail", [][2]string{
		{"up/down j/k", "Scroll"},
		{"pgup/pgdn", "Page scroll"},
		{"home/end", "Top/bottom"},
		{"tab or t", "Toggle block"},
		{"e", "Export as Markdown"},
		{"esc", "Back to list"},
	})

	section("Search", [][2]string{
		{"enter", "Execute search"},
		{"up/down", "Navigate results"},
		{"enter", "Open result"},
		{"/", "Focus search input"},
		{"esc", "Back"},
	})

	section("Stats", [][2]string{
		{"d", "Daily view"},
		{"w", "Weekly view"},
		{"m", "Monthly view"},
		{"up/down", "Scroll"},
		{"esc", "Back"},
	})

	return b.String()
}

func (v *HelpView) StatusKeys() string {
	return "?:close help"
}
