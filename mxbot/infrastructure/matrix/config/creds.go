package config

// BotCredentials - credentials of the bot for Matrix
// should be provided by the user
// (username, password, homeserverURL)
type BotCredentials struct {
	Username        string
	Password        string
	HomeServerURL   string
	PickleKey       []byte
	IsThreadEnabled bool //????
	ThreadLimit     int
	// AuthMode        AuthMode
	// ASToken         string
}
