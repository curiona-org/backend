package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/curiona-org/backend/internal/api/websocket"
	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/llm"
)

func (a *API) RoadmapChatAssist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	roadmap, err := a.application.GetRoadmapBySlug(ctx, slug)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	roadmapBaseKnowledge := io.StreamRoadmapLLMInput{
		GetRoadmapOutput: roadmap,
	}

	client, err := a.ws.Handle(w, r)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	auth := auth.FromContext(ctx)
	roomName := fmt.Sprintf("roadmap:%d:%d", roadmap.ID, auth.AccountID)

	client.SetRoom(roomName)
	a.ws.AddClient(client)

	go func() {
		// Send welcome message
		welcomeMessage := websocket.NewMessageEvent{
			Message: "👋 Welcome to the Roadmap Chat Assist! We're here to help you with your roadmap. Let's get started! 🚀 Feel free to ask any questions you have.",
			From:    "Curiona 🤖",
			Sent:    time.Now(),
		}

		data, err := json.Marshal(welcomeMessage)
		if err != nil {
			log.Err(err).Msg("error marshalling welcome message")
			return
		}

		client.WriteDirectMessage(websocket.Event{
			Type:    websocket.EventNewMessage,
			Payload: data,
		})
	}()

	a.ws.RegisterEventHandler(websocket.EventRoadmapChatAssistRequest, func(event websocket.Event, client *websocket.Client) error {
		var chatAssistEvent websocket.RoadmapChatAssistRequestEvent
		if err := json.Unmarshal(event.Payload, &chatAssistEvent); err != nil {
			return cerrors.ErrInvalidData
		}

		go func() {
			roadmapBaseKnowledge.Message = chatAssistEvent.Message
			llmStream, err := a.application.StreamRoadmapLLM(ctx, roadmapBaseKnowledge)
			if err != nil {
				return
			}

			for {
				content, err := llmStream.Recv()
				if err != nil && errors.Is(err, llm.StreamDone) {
					doneChunk := websocket.RoadmapChatAssistChunkEvent{Content: "", Done: true}

					var data []byte
					data, err = json.Marshal(doneChunk)
					if err != nil {
						log.Err(err).Msg("error marshalling message, sending literal json string")
						data = []byte(`{"content":"","done":true}`)
					}

					client.WriteDirectMessage(websocket.Event{
						Type:    websocket.EventRoadmapChatAssistChunk,
						Payload: data,
					})

					break
				}

				if err != nil {
					log.Err(err).Msg("error receiving message from LLM")
					break
				}

				chunk := websocket.RoadmapChatAssistChunkEvent{
					Content: content,
					Done:    false,
				}

				var data []byte
				data, err = json.Marshal(chunk)
				if err != nil {
					// in case of error while marshalling, we'll send a literal json string
					log.Err(err).Msg("error marshalling message, sending literal json string")
					data = append(data, []byte(`{"content":"`)...)
					data = append(data, []byte(content)...)
					data = append(data, []byte(`","done":false}`)...)
				}

				client.WriteDirectMessage(websocket.Event{
					Type:    websocket.EventRoadmapChatAssistChunk,
					Payload: data,
				})
			}
		}()

		return nil
	})

	go client.Read(ctx)
	go client.WriteLoop(ctx)

	select {}
}
