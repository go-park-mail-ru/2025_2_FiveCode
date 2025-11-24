package websocket

import (
	"sync"

	"github.com/rs/zerolog"
)

type Hub struct {
	rooms map[int]map[*Client]bool

	register chan *Client

	unregister chan *Client

	subscribe chan *Client

	mu sync.RWMutex

	logger *zerolog.Logger
}

func NewHub(logger *zerolog.Logger) *Hub {
	return &Hub{
		rooms:      make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		subscribe:  make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.logger.Info().
				Int("user_id", client.UserID).
				Msg("client registered")

		case client := <-h.subscribe:
			h.mu.Lock()
			if h.rooms[client.NoteID] == nil {
				h.rooms[client.NoteID] = make(map[*Client]bool)
			}
			h.rooms[client.NoteID][client] = true
			h.mu.Unlock()

			h.logger.Info().
				Int("user_id", client.UserID).
				Int("note_id", client.NoteID).
				Int("clients_in_room", len(h.rooms[client.NoteID])).
				Msg("client subscribed to note")

		case client := <-h.unregister:
			if client.NoteID != 0 {
				h.mu.Lock()
				if clients, ok := h.rooms[client.NoteID]; ok {
					if _, ok := clients[client]; ok {
						delete(clients, client)
						close(client.send)

						if len(clients) == 0 {
							delete(h.rooms, client.NoteID)
						}

						h.logger.Info().
							Int("user_id", client.UserID).
							Int("note_id", client.NoteID).
							Int("clients_left", len(clients)).
							Msg("client unsubscribed from note")
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

func (h *Hub) BroadcastToNote(noteID int, message []byte, excludeUserID int) {
	h.mu.RLock()
	clients := h.rooms[noteID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	for client := range clients {
		if client.UserID == excludeUserID {
			continue
		}

		select {
		case client.send <- message:
		default:
			h.unregister <- client
		}
	}

	h.logger.Debug().
		Int("note_id", noteID).
		Int("recipients", len(clients)-1).
		Msg("message broadcast to note")
}

func (h *Hub) GetClientsInRoom(noteID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[noteID]; ok {
		return len(clients)
	}
	return 0
}
