package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fwznbg/go-multiplayer-tictactoe/components"
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
	// Type    MessageType `json:"type"`
	RoomID  string `json:"roomId"`
	Content []byte `json:"content"`
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

			p.Conn.WriteMessage(websocket.TextMessage, msg.Content)
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
		fmt.Println("status ", h.Room[p.RoomID].Status)
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
				return
			}
		}
		var htmxReq HTMXRequest
		fmt.Println(htmxReq.Move)
		if err := json.Unmarshal(msg, &htmxReq); err != nil {
			// TODO: change to html
			// message := &Message{
			// 	RoomID:  p.RoomID,
			// 	Type:    MessageError,
			// 	Content: "Failed to make move, error unmarshalling",
			// }
			// p.Conn.WriteJSON(message)
			fmt.Println(err)
			return
		}

		x, y, found := strings.Cut(htmxReq.Move, ";")
		if !found {
			// TODO: return html
			return
		}

		intX, err := strconv.Atoi(x)
		if err != nil {
			// TODO: return html
			return
		}
		intY, err := strconv.Atoi(y)
		if err != nil {
			// TODO: return html
			return
		}
		winner, err := p.MakeMove(h, intX, intY)
		if err != nil {
			log.Println(err)
			return
		}

		h.Room[p.RoomID].Winner = winner
		// jsonRoom, err := json.Marshal(h.Room[p.RoomID])
		// if err != nil {
		// 	// TODO: change to html
		// 	message := &Message{
		// 		RoomID:  p.RoomID,
		// 		Type:    MessageError,
		// 		Content: "Failed to marshal room",
		// 	}
		// 	p.Conn.WriteJSON(message)
		// 	return
		// }
		boardWriter := NewBytesWriter()
		components.Board(h.Room[p.RoomID].Board).Render(context.Background(), boardWriter)
		message := &Message{
			RoomID:  p.RoomID,
			Content: boardWriter.Bytes(),
		}

		// if winner == nil {
		// 	message.Type = MessageGameUpdate
		// } else {
		// 	message.Type = MessageGameEnded
		// }

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

func (p *Player) MakeMove(hub *Hub, x, y int) (*Player, error) {
	room := hub.Room[p.RoomID]

	if room.Status != StatusPlaying {
		// TODO: return html
		// msg := &Message{
		// 	Type:    MessageError,
		// 	RoomID:  p.RoomID,
		// 	Content: "The game already ended",
		// }
		// p.Conn.WriteJSON(msg)
		return nil, errors.New("game ended")
	}

	if p.Role == PlayerRoleX && room.Turn != TurnX {
		// TODO: return html
		// msg := &Message{
		// 	Type:    MessageError,
		// 	RoomID:  p.RoomID,
		// 	Content: "Not your move",
		// }
		// p.Conn.WriteJSON(msg)
		return nil, errors.New("not your move")
	} else if p.Role == PlayerRoleO && room.Turn != TurnO {
		// TODO: return html
		// msg := &Message{
		// 	Type:    MessageError,
		// 	RoomID:  p.RoomID,
		// 	Content: "Not your move",
		// }
		// p.Conn.WriteJSON(msg)
		return nil, errors.New("not your move")
	}

	if x < 0 || y < 0 || x > 2 || y > 2 {
		// TODO: return html
		// msg := &Message{
		// 	Type:    MessageError,
		// 	RoomID:  p.RoomID,
		// 	Content: "Invalid move",
		// }
		// p.Conn.WriteJSON(msg)
		return nil, errors.New("invalid move")
	}

	if room.Board[x][y] != "" {
		// TODO: return html
		// msg := &Message{
		// 	Type:    MessageError,
		// 	RoomID:  p.RoomID,
		// 	Content: "Point already occupied",
		// }
		// p.Conn.WriteJSON(msg)
		return nil, errors.New("already occupied")
	}

	room.Board[x][y] = p.GetRole()
	room.SwitchTurn()

	winner := room.CheckWinner()
	if winner != nil {
		room.Status = StatusEnded
	}
	fmt.Println(room.Board)
	return winner, nil
}
