package model

import "time"

// MessageType enumerates the top-level JSONL line types.
type MessageType string

const (
	MsgTypeUser                MessageType = "user"
	MsgTypeAssistant           MessageType = "assistant"
	MsgTypeSystem              MessageType = "system"
	MsgTypeProgress            MessageType = "progress"
	MsgTypeSummary             MessageType = "summary"
	MsgTypeFileHistorySnapshot MessageType = "file-history-snapshot"
)

// SystemSubtype further categorizes system messages.
type SystemSubtype string

const (
	SubtypeAPIError     SystemSubtype = "api_error"
	SubtypeTurnDuration SystemSubtype = "turn_duration"
)

// ContentBlockType enumerates content block types within a message.
type ContentBlockType string

const (
	BlockText       ContentBlockType = "text"
	BlockThinking   ContentBlockType = "thinking"
	BlockToolUse    ContentBlockType = "tool_use"
	BlockToolResult ContentBlockType = "tool_result"
)

// ContentBlock is a single block within message.content (when content is an array).
type ContentBlock struct {
	Type ContentBlockType `json:"type"`

	// For "text" blocks
	Text string `json:"text,omitempty"`

	// For "thinking" blocks
	Thinking string `json:"thinking,omitempty"`

	// For "tool_use" blocks
	ToolUseID string         `json:"id,omitempty"`
	ToolName  string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`

	// For "tool_result" blocks
	ToolUseIDRef string `json:"tool_use_id,omitempty"`
	// Content can be string or array — handled via custom unmarshal
	ResultContent string `json:"-"`
}

// Usage tracks token usage for an assistant message.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// Message is a parsed JSONL line from a session file.
type Message struct {
	Type        MessageType   `json:"-"`
	Subtype     SystemSubtype `json:"-"`
	UUID        string        `json:"-"`
	ParentUUID  string        `json:"-"`
	SessionID   string        `json:"-"`
	Timestamp   time.Time     `json:"-"`
	IsSidechain bool          `json:"-"`

	// For user/assistant messages
	Role        string         // "user" or "assistant"
	ContentText string         // Simple string content (user messages)
	Content     []ContentBlock // Array content (assistant messages)
	Model       string         // Model ID from assistant messages
	StopReason  string         // "end_turn", "max_tokens", etc.
	UsageData   *Usage         // Token usage (assistant only)

	// For system messages
	DurationMs int64  // turn_duration subtype
	ErrorCode  string // api_error code
	ErrorInfo  string // api_error description

	// For summary messages
	SummaryText string
}

// DisplayText returns the primary text content of a message for display.
func (m *Message) DisplayText() string {
	if m.ContentText != "" {
		return m.ContentText
	}
	var text string
	for _, b := range m.Content {
		if b.Type == BlockText {
			if text != "" {
				text += "\n"
			}
			text += b.Text
		}
	}
	return text
}

// ToolUseBlocks returns all tool_use blocks from the message content.
func (m *Message) ToolUseBlocks() []ContentBlock {
	var blocks []ContentBlock
	for _, b := range m.Content {
		if b.Type == BlockToolUse {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

// ThinkingBlocks returns all thinking blocks from the message content.
func (m *Message) ThinkingBlocks() []ContentBlock {
	var blocks []ContentBlock
	for _, b := range m.Content {
		if b.Type == BlockThinking {
			blocks = append(blocks, b)
		}
	}
	return blocks
}
