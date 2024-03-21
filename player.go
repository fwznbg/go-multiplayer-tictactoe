package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type PlayerRoleType int
type MessageType int

const (
	PlayerRoleX PlayerRoleType = iota
	PlayerRoleO
)

const (
	MessageInfo MessageType = iota
	MessageError
	MessageGameEnded
	MessageGameUpdate
)

type Player struct {
	Conn    *websocket.Conn `json:"-"`
	Role    PlayerRoleType  `json:"role"`
	RoomID  string          `json:"roomId"`
	Message chan *Message   `json:"-"`
}

type Message struct {
	Type    MessageType `json:"type"`
	RoomID  string      `json:"roomId"`
	Content string      `json:"content"`
}

type PlayerMove struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func HandleWs(roomID string, h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if room, ok := h.Room[roomID]; ok {
		newPlayer := &Player{
			Conn:    conn,
			RoomID:  room.ID,
			Message: make(chan *Message),
		}

		h.Register <- newPlayer
		log.Println("registering player to room ", room.ID)
		// todo
		go newPlayer.WriteMessage()
		newPlayer.ReadMessage(h)
	}
}

func (p *Player) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.Message:
			if !ok {
				return
			}

			p.Conn.WriteJSON(msg)
		case <-ticker.C:
			// todo
		}
	}
}
func (p *Player) ReadMessage(h *Hub) {
	defer func() {
		h.Unregister <- p
		p.Conn.Close()
	}()

	for {
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
				return
			}
		}

		var playerMove PlayerMove
		if err := json.Unmarshal(msg, &playerMove); err != nil {
			message := &Message{
				RoomID:  p.RoomID,
				Type:    MessageError,
				Content: "Failed to make move, error unmarshalling",
			}
			p.Conn.WriteJSON(message)
			return
		}

		winner := p.MakeMove(h, playerMove.X, playerMove.Y)
		h.Room[p.RoomID].Winner = winner
		jsonRoom, err := json.Marshal(h.Room[p.RoomID])
		if err != nil {
			message := &Message{
				RoomID:  p.RoomID,
				Type:    MessageError,
				Content: "Failed to marshal room",
			}
			p.Conn.WriteJSON(message)
			return
		}
		message := &Message{
			RoomID:  p.RoomID,
			Content: string(jsonRoom),
		}

		if winner == nil {
			message.Type = MessageGameUpdate
		} else {
			message.Type = MessageGameEnded
		}

		h.Broadcast <- message
	}
}

func (p Player) GetRole() string {
	role := "X"
	if p.Role == PlayerRoleO {
		role = "O"
	}

	return role
}

func (p *Player) MakeMove(hub *Hub, x, y int) *Player {
	room := hub.Room[p.RoomID]

	if room.Status != StatusPlaying {
		msg := &Message{
			Type:    MessageError,
			RoomID:  p.RoomID,
			Content: "The game already ended",
		}
		p.Conn.WriteJSON(msg)
		return nil
	}

	if p.Role == PlayerRoleX && room.Turn != TurnX {
		msg := &Message{
			Type:    MessageError,
			RoomID:  p.RoomID,
			Content: "Not your move",
		}
		p.Conn.WriteJSON(msg)
		return nil
	} else if p.Role == PlayerRoleO && room.Turn != TurnO {
		msg := &Message{
			Type:    MessageError,
			RoomID:  p.RoomID,
			Content: "Not your move",
		}
		p.Conn.WriteJSON(msg)
		return nil
	}

	if x < 0 || y < 0 || x > 2 || y > 2 {
		msg := &Message{
			Type:    MessageError,
			RoomID:  p.RoomID,
			Content: "Invalid move",
		}
		p.Conn.WriteJSON(msg)
		return nil
	}

	if room.Board[x][y] != "" {
		msg := &Message{
			Type:    MessageError,
			RoomID:  p.RoomID,
			Content: "Point already occupied",
		}
		p.Conn.WriteJSON(msg)
		return nil
	}

	room.Board[x][y] = p.GetRole()
	room.SwitchTurn()

	winner := room.CheckWinner()
	if winner != nil {
		room.Status = StatusEnded
	}
	return winner
}
