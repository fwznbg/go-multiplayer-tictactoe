package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/fwznbg/go-multiplayer-tictactoe/components"
	"github.com/gorilla/mux"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	components.Home().Render(context.Background(), w)
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomCode := vars["room"]

	components.Room(roomCode).Render(context.Background(), w)
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

	resp := CreateRoomResponse{
		Code: code,
	}

	jsonResp, err := json.Marshal(&resp)
	if err != nil {
		http.Error(w, "failed to encode json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonResp)
}

func startGameHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomCode := vars["room"]

	if len(roomCode) < 6 {
		http.Error(w, "invalid code", http.StatusNotFound)
		return
	}

	if _, ok := hub.Room[roomCode]; !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	HandleWs(roomCode, hub, w, r)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not found"))
}

func main() {
	hub := NewHub()

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/{room}", roomHandler).Methods("GET")

	r.HandleFunc("/api/create-room", func(w http.ResponseWriter, r *http.Request) {
		createRoomHandler(hub, w, r)
	}).Methods("GET")
	r.HandleFunc("/api/{room}", func(w http.ResponseWriter, r *http.Request) {
		startGameHandler(hub, w, r)
	})
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	go hub.Run()
	log.Fatal(http.ListenAndServe(":8080", r))
}
