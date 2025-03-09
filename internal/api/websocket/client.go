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

func NewClient(conn *websocket.Conn, manager *Manager, room string) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Event),
		room:    room,
	}
}

func (c *Client) Read(ctx context.Context) {
	defer func() {
		c.manager.Unregister(c)
	}()

	log := logger.Get()

	c.conn.SetReadLimit(512)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
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

func (c *Client) Write(ctx context.Context, fn func(conn Connection, message Event) error) {
	log := logger.Get()

	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.Unregister(c)
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

			if err := fn(c.conn, message); err != nil {
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

func (c *Client) pongHandler(msg string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
