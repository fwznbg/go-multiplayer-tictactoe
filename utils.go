package main

import (
	"context"

	"github.com/fwznbg/go-multiplayer-tictactoe/components"
)

func GetNotificationComponent(message string, isError bool) []byte{
	writer := NewBytesWriter()
	components.Notification(message, isError).Render(context.Background(), writer)

	return writer.Bytes()
}

func HideNotificationComponent() []byte{
	writer := NewBytesWriter()
	components.HideNotification().Render(context.Background(), writer)

	return writer.Bytes()
}

func GetBoardComponent(board [3][3]string, role, turn int) []byte{
	writer := NewBytesWriter()
	components.Board(board, role, turn).Render(context.Background(), writer)

	return writer.Bytes()
}