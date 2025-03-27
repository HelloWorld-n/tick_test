package types

type Message struct {
	From    string      `json:"From"`
	To      string      `json:"To"`
	When    ISO8601Date `json:"When"`
	Content string      `json:"Content"`
}

type MessageToSend struct {
	Message Message `json:"Message"`
}
