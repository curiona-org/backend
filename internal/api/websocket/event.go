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
	EventSendMessage              = "send_message"
	EventNewMessage               = "new_message"
	EventRoadmapChatAssistRequest = "roadmap_chat_assist_request"
	EventRoadmapChatAssistChunk   = "roadmap_chat_assist_chunk"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

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

func SendMessageHandler(event Event, client *Client) error {
	var chatEvent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return cerrors.ErrInvalidData
	}

	broadMessage := NewMessageEvent{
		Message: chatEvent.Message,
		From:    chatEvent.From,
		Sent:    time.Now(),
	}

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return err
	}

	outgoingEvent := Event{
		Type:    EventNewMessage,
		Payload: data,
	}

	for c := range client.manager.clients {
		if c.room == client.room {
			c.egress <- outgoingEvent
		}
	}

	return nil
}
