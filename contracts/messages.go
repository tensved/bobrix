package contracts

type ChatRole string

const (
	// UserRole represents a regular user in the chat
	UserRole ChatRole = "user"
	// AssistantRole represents an AI assistant in the chat
	AssistantRole ChatRole = "assistant"
)

type Messages []map[ChatRole]string
