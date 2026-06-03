package main

import (
	output "fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Lobby struct {
	HostName string
	Players  []string
}

var lobbiesList = []Lobby{}

func InitSocketListening(socket *websocket.Conn) {
	defer socket.Close()
	for {
		var incomingData map[string]any

		var err = socket.ReadJSON(&incomingData)
		if err != nil {
			println("Player disconnected or sent broken data:", err)
			break
		}

		socket.WriteJSON(map[string]any{"Mf": "yo"})
	}
}

func main() {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/wss", func(w http.ResponseWriter, r *http.Request) {
		var socket, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			output.Println("Conversion of HTTP to WebSocket failed.", err)
			return
		}
		go InitSocketListening(socket)
	})

	// 2. Start the master listener loop at the absolute bottom of main
	output.Println("Server booting up and locking down port :8080...")

	// Notice the colon prefix (":8080"). This halts main() and keeps it listening forever.
	var listenErr = http.ListenAndServe(":"+port, nil)
	if listenErr != nil {
		output.Println("The server crashed during initialization for port "+port+":", listenErr)
	}
}
