package types

type Message struct {
	From    string      `json:"from"`
	To      string      `json:"to"`
	When    ISO8601Date `json:"when"`
	Content string      `json:"content"`
}

type MessageToSend struct {
	Message Message `json:"message"`
}
