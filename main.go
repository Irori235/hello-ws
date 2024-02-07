package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type server struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewServer() *server {
	return &server{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *server) sendToEveryone(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		err := websocket.Message.Send(conn, msg)
		if err != nil {
			log.Println("Broadcast Error:", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *server) webSocketHandler(ws *websocket.Conn) {
	defer ws.Close()

	s.mu.Lock()
	s.clients[ws] = true
	s.mu.Unlock()

	err := websocket.Message.Send(ws, "Server: Hello, Client!")
	if err != nil {
		log.Println(err)
	}

	joinMsg := fmt.Sprintf("Server: New Member joined. (Total: %d)", len(s.clients))

	s.sendToEveryone(joinMsg)

	for {
		msg := ""
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Println(err)
			break
		}

		broadcastMsg := fmt.Sprintf("Server: '%s' received.", msg)
		s.sendToEveryone(broadcastMsg)
	}
}

func main() {
	srv := NewServer()

	http.Handle("/", http.FileServer(http.Dir("./tmp")))
	http.Handle("/ws", websocket.Handler(srv.webSocketHandler))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
