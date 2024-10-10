package contracts

type ChatRole string

const AssistantRole ChatRole = "assistant"
const UserRole ChatRole = "user"

type Messages []map[ChatRole]string
