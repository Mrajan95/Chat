// server.go
package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader to upgrade HTTP connections to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Keep track of all clients
var clients = make(map[*websocket.Conn]bool)

// Channel to broadcast messages to all clients
var broadcast = make(chan interface{}) // Updated: generic type

func main() {
	// Serve the HTML file
	http.HandleFunc("/", serveHome)

	// Handle WebSocket connections
	http.HandleFunc("/ws", handleConnections)

	// Start broadcasting messages to all clients
	go handleMessages()

	// Start server on port 8080
	fmt.Println("Server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Server Error: " + err.Error())
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer ws.Close()

	// Register new client
	clients[ws] = true

	for {
		var msg interface{} // Updated: handle both strings and base64 image strings
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Read Error:", err)
			delete(clients, ws)
			break
		}
		// Send message to broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}
