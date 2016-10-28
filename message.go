package ledgertools

// Message represents a single email message
type Message struct {
	Date      string
	To        string
	From      string
	Subject   string
	TextPlain string
}

// NewMessage creates a new message
func NewMessage(date, to, from, subject, textPlain string) Message {
	return Message{date, to, from, subject, textPlain}
}
