package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/model"
	"github.com/ccss/internal/theme"
)

// SessionDetailView displays a full conversation.
type SessionDetailView struct {
	viewport viewport.Model
	session  model.SessionEntry
	messages []*model.Message
	ready    bool

	// Track which tool/thinking blocks are expanded (by index in the flat block list)
	expandedBlocks map[int]bool
	blockPositions []int // line number where each collapsible block starts
	blockCount     int

	width, height int
}

func NewSessionDetailView() *SessionDetailView {
	return &SessionDetailView{
		expandedBlocks: make(map[int]bool),
	}
}

func (v *SessionDetailView) SetSize(w, h int) {
	v.width = w
	v.height = h
	if v.ready {
		v.viewport.Width = w
		v.viewport.Height = h - 2
	}
}

// SetSession loads a session's messages into the detail view.
func (v *SessionDetailView) SetSession(session model.SessionEntry, messages []*model.Message) {
	v.session = session
	v.messages = messages
	v.expandedBlocks = make(map[int]bool)
	v.viewport = viewport.New(v.width, v.height-2)
	v.viewport.SetContent(v.renderContent())
	v.viewport.GotoTop()
	v.ready = true
}

func (v *SessionDetailView) Update(msg tea.Msg) tea.Cmd {
	if !v.ready {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "t":
			v.toggleNextBlock()
			v.viewport.SetContent(v.renderContent())
			return nil
		case "home", "g":
			v.viewport.GotoTop()
			return nil
		case "end", "G":
			v.viewport.GotoBottom()
			return nil
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

func (v *SessionDetailView) toggleNextBlock() {
	if v.blockCount == 0 {
		return
	}
	currentLine := v.viewport.YOffset
	closest := 0
	minDist := 999999
	for i, pos := range v.blockPositions {
		dist := pos - currentLine
		if dist < 0 {
			dist = -dist
		}
		if dist < minDist {
			minDist = dist
			closest = i
		}
	}
	v.expandedBlocks[closest] = !v.expandedBlocks[closest]
}

func (v *SessionDetailView) renderContent() string {
	var b strings.Builder
	v.blockPositions = nil
	v.blockCount = 0
	lineCount := 0

	boxWidth := v.width - 6
	if boxWidth < 20 {
		boxWidth = 20
	}
	innerWidth := boxWidth - 4 // padding inside box

	// Session metadata header
	b.WriteString("\n")
	lineCount++

	metaParts := []string{v.session.ProjectPath}
	metaParts = append(metaParts, fmt.Sprintf("%d messages", len(v.messages)))
	metaParts = append(metaParts, v.session.Modified.Local().Format("2006-01-02 15:04"))
	if v.session.GitBranch != "" && v.session.GitBranch != "HEAD" {
		metaParts = append(metaParts, v.session.GitBranch)
	}

	metaBox := lipgloss.NewStyle().
		Foreground(theme.TextDim).
		PaddingLeft(2).
		Render("  " + strings.Join(metaParts, "  ·  "))
	b.WriteString(metaBox + "\n\n")
	lineCount += 2

	for _, msg := range v.messages {
		switch msg.Type {
		case model.MsgTypeUser:
			text := msg.ContentText
			if text == "" {
				text = msg.DisplayText()
			}

			// Build inner content
			var inner strings.Builder
			if text != "" {
				inner.WriteString(wrapText(text, innerWidth))
			}

			// Tool results from user messages
			for _, block := range msg.Content {
				if block.Type == model.BlockToolResult {
					blockIdx := v.blockCount
					v.blockPositions = append(v.blockPositions, lineCount)
					v.blockCount++

					if v.expandedBlocks[blockIdx] {
						inner.WriteString("\n")
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.Orange).Bold(true).Render("▼ Tool Result") + "\n")
						result := block.ResultContent
						if len(result) > 2000 {
							result = result[:2000] + "..."
						}
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.TextSec).Render(wrapText(result, innerWidth-2)))
					} else {
						summary := truncate(strings.ReplaceAll(block.ResultContent, "\n", " "), 50)
						inner.WriteString("\n")
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.Orange).Render("▶ Tool Result") +
							lipgloss.NewStyle().Foreground(theme.TextDim).Render(" "+summary))
					}
				}
			}

			// Render user box
			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Blue).
				Width(boxWidth).
				Padding(0, 1).
				Render(
					lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render("User") + "\n" +
						inner.String())
			b.WriteString("  " + box + "\n\n")
			lineCount += strings.Count(box, "\n") + 3

		case model.MsgTypeAssistant:
			// Build assistant header
			modelInfo := ""
			if msg.Model != "" {
				modelInfo = lipgloss.NewStyle().Foreground(theme.TextDim).Render(" " + shortModel(msg.Model))
			}
			header := lipgloss.NewStyle().Foreground(theme.Green).Bold(true).Render("Assistant") + modelInfo

			// Build inner content
			var inner strings.Builder
			for _, block := range msg.Content {
				switch block.Type {
				case model.BlockText:
					if inner.Len() > 0 {
						inner.WriteString("\n")
					}
					inner.WriteString(wrapText(block.Text, innerWidth))

				case model.BlockThinking:
					blockIdx := v.blockCount
					v.blockPositions = append(v.blockPositions, lineCount)
					v.blockCount++
					thinkText := block.Text
					if thinkText == "" {
						thinkText = block.Thinking
					}

					inner.WriteString("\n")
					if v.expandedBlocks[blockIdx] {
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.Pink).Bold(true).Render("▼ Thinking") + "\n")
						text := thinkText
						if len(text) > 3000 {
							text = text[:3000] + "..."
						}
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.TextSec).Render(wrapText(text, innerWidth-2)))
					} else {
						preview := truncate(strings.ReplaceAll(thinkText, "\n", " "), 50)
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.Pink).Render("▶ Thinking") +
							lipgloss.NewStyle().Foreground(theme.TextDim).Render(" "+preview))
					}

				case model.BlockToolUse:
					blockIdx := v.blockCount
					v.blockPositions = append(v.blockPositions, lineCount)
					v.blockCount++

					inner.WriteString("\n")
					if v.expandedBlocks[blockIdx] {
						inner.WriteString(lipgloss.NewStyle().Foreground(theme.Orange).Bold(true).
							Render("▼ "+block.ToolName) + "\n")
						if block.Input != nil {
							inputJSON, _ := json.MarshalIndent(block.Input, "  ", "  ")
							inputStr := string(inputJSON)
							if len(inputStr) > 2000 {
								inputStr = inputStr[:2000] + "..."
							}
							inner.WriteString(lipgloss.NewStyle().Foreground(theme.TextSec).Render(inputStr))
						}
					} else {
						summary := ""
						if block.Input != nil {
							summary = toolInputSummary(block.ToolName, block.Input)
						}
						toolLine := lipgloss.NewStyle().Foreground(theme.Orange).Render("▶ " + block.ToolName)
						if summary != "" {
							toolLine += lipgloss.NewStyle().Foreground(theme.TextDim).Render(" " + summary)
						}
						inner.WriteString(toolLine)
					}
				}
			}

			// Token usage line inside box
			if msg.UsageData != nil {
				usageStr := fmt.Sprintf("%d in / %d out",
					msg.UsageData.InputTokens, msg.UsageData.OutputTokens)
				if msg.UsageData.CacheReadInputTokens > 0 {
					usageStr += fmt.Sprintf(" / %d cached", msg.UsageData.CacheReadInputTokens)
				}
				inner.WriteString("\n" + lipgloss.NewStyle().Foreground(theme.TextDim).Render(usageStr))
			}

			// Render assistant box
			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Green).
				Width(boxWidth).
				Padding(0, 1).
				Render(header + "\n" + inner.String())
			b.WriteString("  " + box + "\n\n")
			lineCount += strings.Count(box, "\n") + 3

		case model.MsgTypeSystem:
			if msg.Subtype == model.SubtypeTurnDuration && msg.DurationMs > 0 {
				secs := float64(msg.DurationMs) / 1000
				var durStr string
				if secs >= 60 {
					durStr = fmt.Sprintf("%dm %.0fs", int(secs)/60, float64(int(secs)%60))
				} else {
					durStr = fmt.Sprintf("%.1fs", secs)
				}
				sep := lipgloss.NewStyle().Foreground(theme.Border).Render(
					"  " + safeRepeat("─", boxWidth/3) +
						lipgloss.NewStyle().Foreground(theme.TextDim).Render(" "+durStr+" ") +
						safeRepeat("─", boxWidth/3))
				b.WriteString(sep + "\n\n")
				lineCount += 2
			} else if msg.Subtype == model.SubtypeAPIError {
				errLine := lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render(
					"  ✗ Error: " + msg.ErrorCode)
				b.WriteString(errLine + "\n")
				lineCount++
			}

		case model.MsgTypeSummary:
			if msg.SummaryText != "" {
				b.WriteString(lipgloss.NewStyle().Foreground(theme.TextDim).Italic(true).Render(
					"  ~ Summary: "+truncate(msg.SummaryText, 80)) + "\n")
				lineCount++
			}
		}
	}

	return b.String()
}

// View renders the detail view.
func (v *SessionDetailView) View() string {
	if !v.ready {
		return "Loading..."
	}
	return v.viewport.View()
}

// StatusKeys returns status bar text for this view.
func (v *SessionDetailView) StatusKeys() string {
	pct := 0
	if v.ready {
		pct = int(v.viewport.ScrollPercent() * 100)
	}
	return fmt.Sprintf("↑↓:scroll  tab:toggle  e:export  esc:back  %d%%", pct)
}

func (v *SessionDetailView) Session() model.SessionEntry {
	return v.session
}

func (v *SessionDetailView) Messages() []*model.Message {
	return v.messages
}

func shortModel(modelID string) string {
	parts := strings.Split(modelID, "-")
	if len(parts) >= 4 && parts[0] == "claude" {
		return parts[1] + "-" + parts[2] + "-" + parts[3]
	}
	return modelID
}

func toolInputSummary(toolName string, input map[string]any) string {
	switch toolName {
	case "Bash":
		if cmd, ok := input["command"]; ok {
			s := fmt.Sprintf("%v", cmd)
			return truncate(strings.ReplaceAll(s, "\n", " "), 50)
		}
	case "Read":
		if fp, ok := input["file_path"]; ok {
			return fmt.Sprintf("%v", fp)
		}
	case "Write":
		if fp, ok := input["file_path"]; ok {
			return fmt.Sprintf("%v", fp)
		}
	case "Edit":
		if fp, ok := input["file_path"]; ok {
			return fmt.Sprintf("%v", fp)
		}
	case "Grep":
		if p, ok := input["pattern"]; ok {
			return fmt.Sprintf("/%v/", p)
		}
	case "Glob":
		if p, ok := input["pattern"]; ok {
			return fmt.Sprintf("%v", p)
		}
	case "Task":
		if d, ok := input["description"]; ok {
			return fmt.Sprintf("%v", d)
		}
	case "WebSearch":
		if q, ok := input["query"]; ok {
			return fmt.Sprintf("%v", q)
		}
	case "WebFetch":
		if u, ok := input["url"]; ok {
			return truncate(fmt.Sprintf("%v", u), 50)
		}
	}
	return ""
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		runes := []rune(line)
		for len(runes) > width {
			result.WriteString(string(runes[:width]))
			result.WriteString("\n")
			runes = runes[width:]
		}
		result.WriteString(string(runes))
		result.WriteString("\n")
	}
	out := result.String()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out
}
