package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type PlayerRoleType int
type MessageType int

const (
	PlayerRoleUnassigned PlayerRoleType = iota
	PlayerRoleX
	PlayerRoleO
)

type Player struct {
	Conn    *websocket.Conn `json:"-"`
	Role    PlayerRoleType  `json:"role"`
	RoomID  string          `json:"roomId"`
	Message chan *Message   `json:"-"`
}

type Message struct {
	RoomID  string `json:"roomId"`
	Content []byte `json:"content"`
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
			Role:    PlayerRoleUnassigned,
		}
		log.Println("registering", room.ID)
		h.Register <- newPlayer
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
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
				return
			}
		}
		var htmxReq HTMXRequest
		if err := json.Unmarshal(msg, &htmxReq); err != nil {
			log.Println(err)

			p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Failed to make move, error unmarshalling", true))

			return
		}

		x, y, found := strings.Cut(htmxReq.Move, ";")
		if !found {
			log.Println("Can't find tile coordinate")

			p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Failed to parse tile coordinate", true))

			return
		}

		intX, err := strconv.Atoi(x)
		if err != nil {
			log.Println("Failed to convert x coordinate to int")

			p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Failed to convert x coordinate to int", true))

			return
		}

		intY, err := strconv.Atoi(y)
		if err != nil {
			log.Println("Failed to convert y coordinate to int")

			p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Failed to convert y coordinate to int", true))

			return
		}

		winner, err := p.MakeMove(h, intX, intY)
		if err != nil {
			log.Println(err)
		} else {
			currentRoom := h.Room[p.RoomID]
			currentRoom.Winner = winner

			if winner != nil {
				currentRoom.Turn = PlayerRoleUnassigned
			}

			for _, player := range []*Player{currentRoom.PlayerX, currentRoom.PlayerO} {
				if winner != nil {
					if winner.Role == player.Role {
						player.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("You win :D", false))
					} else {
						player.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("You lose :(", false))
					}

				}
				player.Conn.WriteMessage(websocket.TextMessage, GetBoardComponent(currentRoom.Board, int(player.Role), int(currentRoom.Turn)))
			}
		}

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
		return nil, errors.New("game ended")
	}

	if p.Role != room.Turn {
		return nil, errors.New("not your move")
	}

	if x < 0 || y < 0 || x > 2 || y > 2 {
		return nil, errors.New("invalid move")
	}

	if room.Board[x][y] != "" {
		return nil, errors.New("already occupied")
	}

	room.Board[x][y] = p.GetRole()
	room.SwitchTurn()

	winner := room.CheckWinner()
	if winner != nil {
		log.Printf("winner is %v", winner.GetRole())
		room.Status = StatusEnded
	}
	return winner, nil
}
