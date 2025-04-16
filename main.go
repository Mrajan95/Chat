package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan BroadcastPayload)

type BroadcastPayload struct {
	Sender  *websocket.Conn
	Message map[string]interface{}
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	fmt.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic("Server error: " + err.Error())
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()
	clients[ws] = true

	username := r.URL.Query().Get("username")

	// Send a message to all clients when a new user joins
	broadcast <- BroadcastPayload{
		Sender: ws,
		Message: map[string]interface{}{
			"type":     "user_joined",
			"username": username,
		},
	}

	// Listen for messages from this user
	for {
		var msg map[string]interface{}
		if err := ws.ReadJSON(&msg); err != nil {
			fmt.Println("Read error:", err)
			delete(clients, ws)
			break
		}
		broadcast <- BroadcastPayload{Sender: ws, Message: msg}
	}

	// Send a message to all clients when a user leaves
	broadcast <- BroadcastPayload{
		Sender: ws,
		Message: map[string]interface{}{
			"type":     "user_left",
			"username": username,
		},
	}
}

func handleMessages() {
	for {
		payload := <-broadcast
		for client := range clients {
			if err := client.WriteJSON(payload.Message); err != nil {
				fmt.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
