package websocket

import (
	"net/http"
	"sync"
)

type Manager struct {
	mtx      sync.RWMutex
	clients  map[*Client]bool
	handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(map[*Client]bool),
		handlers: make(map[string]EventHandler),
	}

	m.setupEventHandlers()
	return m
}

func (m *Manager) Handle(w http.ResponseWriter, r *http.Request, room string) (*Client, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewClient(conn, m, room), nil
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
}

func (m *Manager) routeEvent(event Event, client *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return ErrEventNotFound
	}

	return handler(event, client)
}

func (m *Manager) Register(client *Client) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.clients[client] = true
}

func (m *Manager) Unregister(client *Client) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.clients[client]; ok {
		client.conn.Close()
		delete(m.clients, client)
	}
}
