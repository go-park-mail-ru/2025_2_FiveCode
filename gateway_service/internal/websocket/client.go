package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

type Client struct {
	conn   *websocket.Conn
	hub    *Hub
	send   chan []byte
	UserID int
	NoteID int
	logger *zerolog.Logger
}

func NewClient(conn *websocket.Conn, hub *Hub, userID int, logger *zerolog.Logger) *Client {
	return &Client{
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, 256),
		UserID: userID,
		NoteID: 0,
		logger: logger,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error().Err(err).Msg("websocket error")
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			c.logger.Error().Err(err).Msg("failed to unmarshal client message")
			c.sendError("Invalid message format")
			continue
		}

		c.handleMessage(&clientMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *ClientMessage) {
	switch msg.Type {

	default:
		c.logger.Warn().
			Str("message_type", string(msg.Type)).
			Int("user_id", c.UserID).
			Int("note_id", c.NoteID).
			Msg("received unsupported message type from client")
		c.sendError("Unsupported message type")
	}
}

func (c *Client) sendError(message string) {
	errorMsg := ServerMessage{
		Type:    MessageTypeError,
		Message: message,
	}

	data, err := json.Marshal(errorMsg)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to marshal error message")
		return
	}

	select {
	case c.send <- data:
	default:
		c.logger.Warn().Msg("client send channel is full")
	}
}

func (c *Client) Send(message []byte) {
	select {
	case c.send <- message:
	default:
		close(c.send)
		c.hub.unregister <- c
	}
}
