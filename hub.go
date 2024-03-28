package main

import (
	"context"

	"github.com/fwznbg/go-multiplayer-tictactoe/components"
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
				if room.PlayerX == nil {
					player.Role = PlayerRoleX
					room.PlayerX = player
				} else if room.PlayerO == nil {
					player.Role = PlayerRoleO
					room.PlayerO = player
				} else {
					// TODO: change the broadcast to return html
					// h.Broadcast <- &Message{
					// 	Type:    MessageError,
					// 	RoomID:  player.RoomID,
					// 	Content: "Room is full",
					// }
				}
				// TODO: change the broadcast to return html
				// h.Broadcast <- &Message{
				// 	Type:    MessageInfo,
				// 	RoomID:  player.RoomID,
				// 	Content: fmt.Sprintf("Player %s joined", player.GetRole()),
				// }

				if room.CheckIsFull() {
					if room.Status == StatusWaiting {
						room.Status = StatusPlaying
					}
					// roomJson, err := json.Marshal(room)
					// if err != nil {
					// 	log.Println("Failed to encode json ", err)
					// 	return
					// }

					boardWriter := NewBytesWriter()
					components.Board(room.Board).Render(context.Background(), boardWriter)
					h.Broadcast <- &Message{
						RoomID:  player.RoomID,
						Content: boardWriter.Bytes(),
					}

					// h.Broadcast <- &Message{
					// 	Type:    MessageGameUpdate,
					// 	RoomID:  player.RoomID,
					// 	Content: string(roomJson),
					// }
				}
			}
		case player := <-h.Unregister:
			if room, ok := h.Room[player.RoomID]; ok {
				if player.Role == PlayerRoleX {
					room.PlayerX = nil
				} else {
					room.PlayerO = nil
				}
				if room.CheckIsEmpty() {
					delete(h.Room, room.ID)
				} else {
					// TODO: change this to return html
					// h.Broadcast <- &Message{
					// 	Type:    MessageInfo,
					// 	RoomID:  room.ID,
					// 	Content: fmt.Sprintf("Player %s is leaving the room", player.GetRole()),
					// }
					room.Status = StatusWaiting
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
