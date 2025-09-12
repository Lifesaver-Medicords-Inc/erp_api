package hub

type Message struct {
	Channel string
	Data    []byte
}

type BaseMessage struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
}

type ChatMessage struct {
}

type OrderMessage struct {
}
