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
		err := c.conn.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close websocket connection")
		}
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Error().Err(err).Msg("failed to set read deadline")
		return
	}
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error().Err(err).Msg("failed to set read deadline in pong handler")
			return err
		}
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
		err := c.conn.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close websocket connection")
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					return
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			if err != nil {
				return
			}

			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write([]byte{'\n'})
				if err != nil {
					return
				}
				_, err = w.Write(<-c.send)
				if err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
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
