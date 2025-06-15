package dto

import "encoding/json"

type SSEMessage struct {
	UserID  string `json:"user_id"`
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

func (m *SSEMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func FromJSON(data []byte) (*SSEMessage, error) {
	var msg SSEMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
