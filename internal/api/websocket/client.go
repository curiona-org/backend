package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn    *websocket.Conn
	manager *Manager
	egress  chan Event
	room    string
}

var (
	pongWait     = time.Minute
	pingInterval = (pongWait * 9) / 10
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Event),
	}
}

func (c *Client) SetRoom(room string) {
	c.room = room
}

func (c *Client) Read(ctx context.Context) {
	defer func() {
		c.manager.RemoveClient(c)
	}()

	log := logger.FromContext(ctx)

	c.conn.SetReadLimit(512)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Err(err).Msg("error reading message")
			}
			break
		}

		var req Event
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Err(err).Msg("error unmarshalling message")
			break
		}

		if err := c.manager.routeEvent(req, c); err != nil {
			log.Err(err).Msg("error handling event")
		}
	}
}

func (c *Client) WriteLoop(ctx context.Context) {
	log := logger.FromContext(ctx)

	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.RemoveClient(c)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Err(err).Msg("error writing close message")
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Err(err).Msg("error marshalling message")
				continue // Don't terminate connection for serialization errors
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				if err == websocket.ErrCloseSent {
					return
				}

				log.Err(err).Msg("error writing message")
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Err(err).Msg("error writing ping")
				return
			}
		}
	}
}

// WriteDirectMessage sends a message to the client without going through the event handler
func (c *Client) WriteDirectMessage(message Event) {
	c.egress <- message
}

func (c *Client) pongHandler(msg string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
