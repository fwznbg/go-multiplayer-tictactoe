package main

import (
	"math/rand"
)

type StatusType int
type TurnType int

const (
	StatusWaiting StatusType = iota
	StatusPlaying
	StatusEnded
)

const (
	TurnO TurnType = iota
	TurnX
)

type GameRoom struct {
	ID      string       `json:"id"`
	PlayerX *Player      `json:"playerX"`
	PlayerO *Player      `json:"playerO"`
	Status  StatusType   `json:"status"`
	Board   [3][3]string `json:"board"`
	Turn    TurnType     `json:"turn"`
	Winner  *Player      `json:"winner"`
}

func NewRoom(roomId string) *GameRoom {
	board := [3][3]string{}
	board[0] = [3]string{"", "", ""}
	board[1] = [3]string{"", "", ""}
	board[2] = [3]string{"", "", ""}

	return &GameRoom{
		ID:      roomId,
		PlayerX: nil,
		PlayerO: nil,
		Status:  StatusWaiting,
		Board:   board,
		Turn:    TurnX,
		Winner:  nil,
	}
}

func GenerateRoom(n int) string {
	letter := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	code := make([]rune, n)

	for i := range code {
		code[i] = letter[rand.Intn(len(letter))]
	}

	return string(code)
}

func (r GameRoom) CheckIsFull() bool {
	return r.PlayerX != nil && r.PlayerO != nil
}

func (r GameRoom) CheckIsEmpty() bool {
	return r.PlayerX == nil && r.PlayerO == nil
}

func (r GameRoom) getOccupier(x, y int) *Player {
	tileValue := r.Board[x][y]
	if tileValue == "" {
		return nil
	}

	if tileValue == "X" {
		return r.PlayerX
	}
	return r.PlayerO
}

func (r GameRoom) CheckWinner() *Player {
	board := r.Board
	var winner *Player

	if board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return r.getOccupier(0, 0)
	}

	if board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return r.getOccupier(0, 2)
	}

	for i, row := range board {
		if row[0] == row[1] && row[1] == row[2] {
			winner = r.getOccupier(i, 0)
			break
		}

		if board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			winner = r.getOccupier(0, i)
			break
		}
	}

	return winner
}

func (r *GameRoom) SwitchTurn() {
	if r.Turn == TurnX {
		r.Turn = TurnO
	} else {
		r.Turn = TurnX
	}
}
