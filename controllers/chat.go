package controllers

import (
	"fmt"
	"net/http"
	"time"
	"web-server/models"
	"web-server/utils"

	"golang.org/x/net/websocket"
)

// Client is the type for all the clients
type Client struct {
	User *models.User
	Conn *websocket.Conn
	Pool *Pool
}

// Message is type of all messages received through Chat
type Message string

// MessageS is type of all messages sent through Chat
type MessageS struct {
	Message `json:"text"`
	Time    *Time        `json:"time"`
	User    *models.User `json:"user"`
	Code    int          `json:"code"`
}

// Read listens to the given client and handles its incoming messages. The user is deregistered from its pool when the connection closes or its session expires
func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var message Message
		if err := websocket.JSON.Receive(c.Conn, &message); err != nil {
			fmt.Println("Can't receive")
			break
		}

		r := c.Conn.Request()
		_, b := utils.GlobalSessions.SessionCheck(r)
		if !b {
			fmt.Println(b)
			websocket.JSON.Send(c.Conn, MessageS{"No session", nil, nil, 10})
			break
		}

		t := getTime()
		c.Pool.Broadcast <- &MessageS{message, t, c.User, 0}
	}
}

// Pool is the type for all the pools
type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan *MessageS
}

// NewPool returns a pointer to a newly created pool
func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *MessageS),
	}
}

// Start is responsible for sending all the messages and important user events to users of a pool
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			t := getTime()
			// fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for c := range pool.Clients {
				websocket.JSON.Send(c.Conn, MessageS{"New User Joined", t, client.User, 1})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			t := getTime()
			// fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for c := range pool.Clients {
				websocket.JSON.Send(c.Conn, MessageS{"User left", t, client.User, 2})
			}
			break
		case messageS := <-pool.Broadcast:
			// fmt.Println("Sending message to all clients in Pool")
			for c := range pool.Clients {
				if err := websocket.JSON.Send(c.Conn, messageS); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

// Chat registers a new user into a pool
func Chat(pool *Pool, ws *websocket.Conn) {
	fmt.Println("WebSocket Endpoint Hit")

	r := ws.Request()
	session, b := utils.GlobalSessions.SessionCheck(r)
	if !b {
		websocket.JSON.Send(ws, MessageS{"No session", nil, nil, 10})
		ws.Close()
		return
	}

	client := &Client{
		Conn: ws,
		Pool: pool,
		User: utils.SessionGetUser(&session, r),
	}

	session.Set("pool", pool)
	pool.Register <- client
	client.Read()
}

// deleteFromPool unregisters user in session from its pool
func deleteFromPool(r *http.Request) {
	session, b := utils.GlobalSessions.SessionCheck(r)
	if !b {
		return
	}
	id := session.Get("id")
	pool := session.Get("pool").(*Pool)
	for client := range pool.Clients {
		if client.User.Id == id {
			pool.Unregister <- client
			client.Conn.Close()
		}
	}
}

// Time is struct for sending time to frontend
type Time struct {
	Month  string `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
}

// getTime returns current time in type Time
func getTime() *Time {
	t := time.Now().Local()
	return &Time{t.Month().String()[:3], t.Day(), t.Hour(), t.Minute()}
}
