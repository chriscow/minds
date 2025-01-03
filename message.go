package minds

import (
	"encoding/json"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleFunction  Role = "function"
	RoleTool      Role = "tool"
	RoleAI        Role = "ai"
	RoleModel     Role = "model"
)

type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`     // For function calls
	Metadata   Metadata   `json:"metadata,omitempty"` // For additional context
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"func_response,omitempty"`
}

func (m Message) TokenCount(tokenizer TokenCounter) (int, error) {
	wholeMsg, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}

	return tokenizer.CountTokens(string(wholeMsg))
}
