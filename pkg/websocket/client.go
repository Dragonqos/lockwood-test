package websocket

import (
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
)

type Client struct {
	conn       *gorilla.Conn
	send       chan any
	requests   chan []byte
	done       chan struct{}
	once       sync.Once
	sessionID  string
	onActivity func()
}

func NewClient(conn *gorilla.Conn, sessionID string, queueSize int) *Client {
	return &Client{
		conn:      conn,
		send:      make(chan any, queueSize),
		requests:  make(chan []byte),
		done:      make(chan struct{}),
		sessionID: sessionID,
	}
}

func (c *Client) ReadLoop(maxMessageSize int, pongWait time.Duration) {
	defer close(c.requests)

	c.conn.SetReadLimit(int64(maxMessageSize))
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		if c.onActivity != nil {
			c.onActivity()
		}
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		select {
		case c.requests <- payload:
		case <-c.done:
			return
		}
	}
}

func (c *Client) WriteLoop(writeWait, pingPeriod time.Duration) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.Close(gorilla.CloseGoingAway, "write deadline failed")
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				c.Close(gorilla.CloseGoingAway, "write failed")
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.Close(gorilla.CloseGoingAway, "write deadline failed")
				return
			}
			if err := c.conn.WriteMessage(gorilla.PingMessage, nil); err != nil {
				c.Close(gorilla.CloseGoingAway, "ping failed")
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) Requests() <-chan []byte { return c.requests }

func (c *Client) SendResponse(rid any, command, status string) {
	c.SendMessage(struct {
		RID    any    `json:"rID"`
		Cmd    string `json:"cmd"`
		Status string `json:"status"`
	}{RID: rid, Cmd: command, Status: status})
}

func (c *Client) SendMessage(message any) {
	select {
	case c.send <- message:
	case <-c.done:
	default:
		c.Close(gorilla.ClosePolicyViolation, "outbound queue full")
	}
}

func (c *Client) Close(code int, reason string) {
	c.once.Do(func() {
		close(c.done)
		_ = c.conn.WriteControl(gorilla.CloseMessage, gorilla.FormatCloseMessage(code, reason), time.Now().Add(10*time.Second))
		_ = c.conn.Close()
	})
}

func (c *Client) SessionID() string { return c.sessionID }

func (c *Client) SetActivityHandler(handler func()) { c.onActivity = handler }
