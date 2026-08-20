package contracts

type ChatRole string

const (
	// UserRole represents a regular user in the chat
	UserRole ChatRole = "user"
	// AssistantRole represents an AI assistant in the chat
	AssistantRole ChatRole = "assistant"
)

// ChatMessage is one turn of chat history. Attachments holds mxc:// URLs of any
// images/audio/video/files the original Matrix message carried; text-only
// handlers can simply ignore it, multimodal ones can download and attach them.
type ChatMessage struct {
	Role        ChatRole
	Content     string
	Attachments []string
}

type Messages []ChatMessage
