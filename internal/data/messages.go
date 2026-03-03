package data

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ccss/internal/model"
)

// LoadSessionMessages reads a JSONL session file and returns parsed messages.
// Skips file-history-snapshot and progress lines. Uses streaming parsing.
func LoadSessionMessages(jsonlPath string) ([]*model.Message, error) {
	normalizedPath := filepath.FromSlash(jsonlPath)

	f, err := os.Open(normalizedPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []*model.Message
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msgType := peekType(line)
		if msgType == "file-history-snapshot" || msgType == "progress" {
			continue
		}

		msg, err := parseMessage(line)
		if err != nil || msg == nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, scanner.Err()
}

// peekType extracts the "type" field value without full JSON parsing.
func peekType(line []byte) string {
	idx := bytes.Index(line, []byte(`"type":"`))
	if idx == -1 {
		return ""
	}
	start := idx + 8
	if start >= len(line) {
		return ""
	}
	end := bytes.IndexByte(line[start:], '"')
	if end == -1 {
		return ""
	}
	return string(line[start : start+end])
}

// parseMessage does full JSON parsing of a JSONL line into a Message.
func parseMessage(line []byte) (*model.Message, error) {
	var raw struct {
		Type        string          `json:"type"`
		Subtype     string          `json:"subtype,omitempty"`
		UUID        string          `json:"uuid"`
		ParentUUID  string          `json:"parentUuid"`
		SessionID   string          `json:"sessionId"`
		Timestamp   string          `json:"timestamp"`
		IsSidechain bool            `json:"isSidechain"`
		Message     json.RawMessage `json:"message"`
		DurationMs  int64           `json:"durationMs,omitempty"`
		Summary     string          `json:"summary,omitempty"`
		Cause       *struct {
			Code string `json:"code"`
		} `json:"cause,omitempty"`
	}
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil, err
	}

	msg := &model.Message{
		Type:        model.MessageType(raw.Type),
		Subtype:     model.SystemSubtype(raw.Subtype),
		UUID:        raw.UUID,
		ParentUUID:  raw.ParentUUID,
		SessionID:   raw.SessionID,
		IsSidechain: raw.IsSidechain,
		DurationMs:  raw.DurationMs,
		SummaryText: raw.Summary,
	}

	if raw.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339Nano, raw.Timestamp); err == nil {
			msg.Timestamp = t
		}
	}

	if raw.Cause != nil {
		msg.ErrorCode = raw.Cause.Code
		msg.ErrorInfo = raw.Cause.Code
	}

	if raw.Message != nil {
		switch msg.Type {
		case model.MsgTypeUser:
			parseUserMessage(raw.Message, msg)
		case model.MsgTypeAssistant:
			parseAssistantMessage(raw.Message, msg)
		}
	}

	return msg, nil
}

func parseUserMessage(raw json.RawMessage, msg *model.Message) {
	var inner struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &inner); err != nil {
		return
	}
	msg.Role = inner.Role

	// Try as string first
	var s string
	if json.Unmarshal(inner.Content, &s) == nil {
		msg.ContentText = s
		return
	}
	// Try as array of content blocks
	var blocks []model.ContentBlock
	if json.Unmarshal(inner.Content, &blocks) == nil {
		msg.Content = blocks
		// Extract tool_result text
		for i, b := range blocks {
			if b.Type == model.BlockToolResult {
				extractToolResultContent(inner.Content, i, &blocks[i])
			}
		}
		msg.Content = blocks
	}
}

func parseAssistantMessage(raw json.RawMessage, msg *model.Message) {
	var inner struct {
		Role       string          `json:"role"`
		Model      string          `json:"model"`
		Content    json.RawMessage `json:"content"`
		StopReason string          `json:"stop_reason"`
		Usage      json.RawMessage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &inner); err != nil {
		return
	}
	msg.Role = inner.Role
	msg.Model = inner.Model
	msg.StopReason = inner.StopReason

	// Parse content blocks
	var blocks []model.ContentBlock
	if json.Unmarshal(inner.Content, &blocks) == nil {
		// Handle thinking blocks that use "thinking" field instead of "text"
		for i, b := range blocks {
			if b.Type == model.BlockThinking && b.Text == "" && b.Thinking != "" {
				blocks[i].Text = b.Thinking
			}
		}
		msg.Content = blocks
	}

	// Parse usage with fallback for different field naming conventions
	if inner.Usage != nil {
		var usage model.Usage
		if json.Unmarshal(inner.Usage, &usage) == nil {
			msg.UsageData = &usage
		}
		// Try alternate field names
		if msg.UsageData != nil && msg.UsageData.CacheReadInputTokens == 0 {
			var altUsage struct {
				CacheReadInputTokens     int `json:"cacheReadInputTokens"`
				CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
			}
			if json.Unmarshal(inner.Usage, &altUsage) == nil {
				if altUsage.CacheReadInputTokens > 0 {
					msg.UsageData.CacheReadInputTokens = altUsage.CacheReadInputTokens
				}
				if altUsage.CacheCreationInputTokens > 0 {
					msg.UsageData.CacheCreationInputTokens = altUsage.CacheCreationInputTokens
				}
			}
		}
	}
}

// extractToolResultContent handles the polymorphic content field in tool_result blocks.
func extractToolResultContent(rawArray json.RawMessage, index int, block *model.ContentBlock) {
	var arr []json.RawMessage
	if json.Unmarshal(rawArray, &arr) != nil || index >= len(arr) {
		return
	}
	var obj struct {
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(arr[index], &obj) != nil {
		return
	}
	// Try as string
	var s string
	if json.Unmarshal(obj.Content, &s) == nil {
		block.ResultContent = s
		return
	}
	// Try as array of text blocks
	var contentBlocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(obj.Content, &contentBlocks) == nil {
		for _, cb := range contentBlocks {
			if cb.Type == "text" {
				block.ResultContent += cb.Text
			}
		}
	}
}
