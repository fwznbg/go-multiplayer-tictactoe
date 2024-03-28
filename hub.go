package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Room       map[string]*GameRoom
	Register   chan *Player
	Unregister chan *Player
	Broadcast  chan *Message
}

func NewHub() *Hub {
	return &Hub{
		Room:       make(map[string]*GameRoom),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
		Broadcast:  make(chan *Message, 2),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case player := <-h.Register:
			if room, ok := h.Room[player.RoomID]; ok {
				if room.CheckIsFull() {
					player.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("The room is already full", true))
				} else {
					if room.PlayerX == nil {
						player.Role = PlayerRoleX
						room.PlayerX = player
						log.Println("registering x")

						if room.PlayerO == nil {
							player.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Waiting for player O", false))
						}
					} else if room.PlayerO == nil {
						player.Role = PlayerRoleO
						room.PlayerO = player
						log.Println("registering o")

						if room.PlayerX == nil {
							player.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Waiting for player X", false))
						}
					}

					if room.CheckIsFull() {
						if room.Status == StatusWaiting {
							room.Status = StatusPlaying
						}

						for _, p := range []*Player{room.PlayerX, room.PlayerO} {
							if room.Turn == p.Role {
								p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Your turn", false))
							} else {
								p.Conn.WriteMessage(websocket.TextMessage, GetNotificationComponent("Enemy's turn", false))
							}

							p.Conn.WriteMessage(websocket.TextMessage, GetBoardComponent(room.Board, int(p.Role), int(room.Turn)))
						}
					}
				}
			}
		case player := <-h.Unregister:
			if room, ok := h.Room[player.RoomID]; ok {
				if player.Role != PlayerRoleUnassigned {
					if player.Role == PlayerRoleX {
						room.PlayerX = nil
					} else {
						room.PlayerO = nil
					}
					if room.CheckIsEmpty() {
						delete(h.Room, room.ID)
					} else {
						h.Broadcast <- &Message{
							RoomID:  player.RoomID,
							Content: HideNotificationComponent(),
						}

						h.Broadcast <- &Message{
							RoomID:  player.RoomID,
							Content: GetNotificationComponent(fmt.Sprintf("Player %s is leaving, waiting them to join again", player.GetRole()), false),
						}
						room.Status = StatusWaiting
					}
				}

				close(player.Message)
			}
		case msg := <-h.Broadcast:
			if room, ok := h.Room[msg.RoomID]; ok {
				if room.PlayerX != nil {
					room.PlayerX.Message <- msg
				}
				if room.PlayerO != nil {
					room.PlayerO.Message <- msg
				}
			}
		}
	}
}
