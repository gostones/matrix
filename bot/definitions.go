package bot

var (
	ProxyUrl  string
	MatrixUrl string
)

// Robot is the interface all robots must follow
type Robot interface {
	Run(*Command) string
	Description() string
}

// Command represents the fields in a ssh-chat-bot command
type Command struct {
	// Command string
	// Args    []string
	ChatMessage
}
