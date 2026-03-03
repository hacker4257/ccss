package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ccss/internal/model"
)

// ExportSessionMarkdown generates a Markdown representation of a session.
func ExportSessionMarkdown(session model.SessionEntry, messages []*model.Message) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Session: %s\n\n", session.Summary))
	buf.WriteString(fmt.Sprintf("- **Project**: %s\n", session.ProjectPath))
	buf.WriteString(fmt.Sprintf("- **Created**: %s\n", session.Created.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("- **Modified**: %s\n", session.Modified.Format("2006-01-02 15:04:05")))
	if session.GitBranch != "" {
		buf.WriteString(fmt.Sprintf("- **Branch**: %s\n", session.GitBranch))
	}
	buf.WriteString(fmt.Sprintf("- **Messages**: %d\n\n", session.MessageCount))
	buf.WriteString("---\n\n")

	for _, msg := range messages {
		switch msg.Type {
		case model.MsgTypeUser:
			if msg.ContentText != "" {
				buf.WriteString(fmt.Sprintf("## 👤 User\n\n%s\n\n", msg.ContentText))
			}

		case model.MsgTypeAssistant:
			header := "## 🤖 Assistant"
			if msg.Model != "" {
				header += fmt.Sprintf(" (%s)", shortModelName(msg.Model))
			}
			buf.WriteString(header + "\n\n")

			for _, block := range msg.Content {
				switch block.Type {
				case model.BlockText:
					buf.WriteString(block.Text + "\n\n")
				case model.BlockThinking:
					text := block.Text
					if text == "" {
						text = block.Thinking
					}
					buf.WriteString("<details>\n<summary>💭 Thinking</summary>\n\n")
					buf.WriteString(text + "\n\n")
					buf.WriteString("</details>\n\n")
				case model.BlockToolUse:
					buf.WriteString(fmt.Sprintf("<details>\n<summary>🔧 Tool: %s</summary>\n\n", block.ToolName))
					if block.Input != nil {
						inputJSON, _ := json.MarshalIndent(block.Input, "", "  ")
						buf.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(inputJSON)))
					}
					buf.WriteString("</details>\n\n")
				}
			}

			if msg.UsageData != nil {
				buf.WriteString(fmt.Sprintf("*Tokens: %d in / %d out*\n\n",
					msg.UsageData.InputTokens, msg.UsageData.OutputTokens))
			}

		case model.MsgTypeSystem:
			if msg.Subtype == model.SubtypeTurnDuration && msg.DurationMs > 0 {
				secs := float64(msg.DurationMs) / 1000
				if secs >= 60 {
					buf.WriteString(fmt.Sprintf("*Turn duration: %dm %.0fs*\n\n",
						int(secs)/60, float64(int(secs)%60)))
				} else {
					buf.WriteString(fmt.Sprintf("*Turn duration: %.1fs*\n\n", secs))
				}
			}
		}
	}

	return buf.String()
}

// WriteExportFile writes markdown content to a file.
func WriteExportFile(outputPath string, content string) error {
	return os.WriteFile(outputPath, []byte(content), 0644)
}

func shortModelName(modelID string) string {
	// "claude-opus-4-6-20260206" -> "opus-4-6"
	parts := strings.Split(modelID, "-")
	if len(parts) >= 4 && parts[0] == "claude" {
		name := parts[1] + "-" + parts[2] + "-" + parts[3]
		return name
	}
	return modelID
}
