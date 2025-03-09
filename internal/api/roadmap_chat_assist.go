package api

import (
	"errors"
	"net/http"

	"github.com/curiona-org/backend/internal/api/websocket"
	"github.com/curiona-org/backend/internal/app/io"
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

	client, err := a.ws.Handle(w, r, slug)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	log.Info().Str("room", slug).Msg("Roadmap chat assist connected")
	a.ws.Register(client)
	log.Info().Str("remote_addr", r.RemoteAddr).Msg("Client registered")

	go client.Read(ctx)

	client.Write(ctx, func(conn websocket.Connection, message websocket.Event) error {
		roadmapBaseKnowledge.Message = string(message.Payload)
		llmStream, err := a.application.StreamRoadmapLLM(ctx, roadmapBaseKnowledge)
		if err != nil {
			return err
		}

		for {
			msg, err := llmStream.Recv()
			if err != nil && errors.Is(err, llm.StreamDone) {
				break
			}

			if err != nil {
				return err
			}

			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return err
			}
		}

		return nil
	})
}
