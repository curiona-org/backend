package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	mtx     sync.RWMutex
	conn    *websocket.Conn
	manager *Manager
	egress  chan Event
	room    string

	// handlers holds a client specific event handler. used for roadmap private chat
	// events since they are not global events.
	handlers map[string]EventHandler
}

var (
	pongWait     = time.Minute
	pingInterval = (pongWait * 9) / 10
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:     conn,
		manager:  manager,
		egress:   make(chan Event),
		handlers: make(map[string]EventHandler),
	}
}

func (c *Client) SetRoom(room string) {
	c.room = room
}

func (c *Client) RegisterEventHandler(event string, handler EventHandler) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.handlers[event] = handler
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

		c.mtx.RLock()
		handler, ok := c.handlers[req.Type]
		c.mtx.RUnlock()

		if ok {
			if err := handler(req, c); err != nil {
				log.Err(err).Msg("error handling event")
			}
			continue
		}

		// If no client specific handler is found, route the event to the global event handler.
		// Warning: This will route the event to all clients in the same room.
		// for example, if a client needs assistant in roadmap X, the event will be routed to all clients in roadmap X if
		// the client and other clients failed to register a client specific handler.
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
