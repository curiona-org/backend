package websocket

import (
	"encoding/json"
	"time"

	"github.com/curiona-org/backend/internal/cerrors"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

var (
	ErrEventNotFound = cerrors.New("event not found")
)

const (
	EventNewMessage               = "new_message"
	EventRoadmapChatAssistRequest = "roadmap_chat_assist_request"
	EventRoadmapChatAssistChunk   = "roadmap_chat_assist_chunk"
)

type NewMessageEvent struct {
	Message string    `json:"message"`
	From    string    `json:"from"`
	Sent    time.Time `json:"sent"`
}

type RoadmapChatAssistRequestEvent struct {
	Message string `json:"message"`
}

type RoadmapChatAssistChunkEvent struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

var RoadmapChatAssistChunkEventDoneJSON = json.RawMessage(`{"content":"","done":true}`)
