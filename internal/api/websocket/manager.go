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

	return m
}

// Handle creates a new client from the request and upgrades the connection to a websocket connection.
// It returns the client and an error if the upgrade fails.
func (m *Manager) Handle(w http.ResponseWriter, r *http.Request) (*Client, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewClient(conn, m), nil
}

func (m *Manager) AddClient(client *Client) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.clients[client] = true
}

func (m *Manager) RemoveClient(client *Client) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.clients[client]; ok {
		client.conn.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) RegisterEventHandler(event string, handler EventHandler) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.handlers[event] = handler
}

func (m *Manager) routeEvent(event Event, client *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return ErrEventNotFound
	}

	return handler(event, client)
}
