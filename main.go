package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/fwznbg/go-multiplayer-tictactoe/components"
	"github.com/gorilla/mux"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	components.Home().Render(context.Background(), w)
}

func roomHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomCode := vars["room"]

	components.Room(roomCode).Render(context.Background(), w)

	if len(roomCode) < 6 {
		components.Notification("Invalid room id", true).Render(context.Background(), w)
		return
	}

	if _, ok := hub.Room[roomCode]; !ok {
		components.Notification("Room id not found", true).Render(context.Background(), w)
		return
	}
}

func createRoomHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	roomExist := true
	var code string

	for roomExist {
		code = GenerateRoom(6)
		if _, ok := hub.Room[code]; !ok {
			roomExist = false
			hub.Room[code] = NewRoom(code)
		}
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", code))
}

func startGameHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomCode := vars["room"]
	HandleWs(roomCode, hub, w, r)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not found"))
}

func main() {
	hub := NewHub()

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/{room}", func(w http.ResponseWriter, r *http.Request) {
		roomHandler(hub, w, r)
	}).Methods("GET")

	r.HandleFunc("/api/create-room", func(w http.ResponseWriter, r *http.Request) {
		createRoomHandler(hub, w, r)
	}).Methods("GET")
	r.HandleFunc("/api/join", func(w http.ResponseWriter, r *http.Request) {
		roomId := r.URL.Query().Get("roomId")
		w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", roomId))
	})
	r.HandleFunc("/api/{room}", func(w http.ResponseWriter, r *http.Request) {
		startGameHandler(hub, w, r)
	})
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	go hub.Run()
	log.Fatal(http.ListenAndServe(":8080", r))
}
