package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

var allowedOrigins = []string{
    "http://localhost:8000",
    "http://yourdomain.com",
    "https://yourdomain.com",
}

func isOriginAllowed(origin string) bool {
    for _, allowedOrigin := range allowedOrigins {
        if strings.HasPrefix(origin, allowedOrigin) {
            return true
        }
    }
    return false
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return isOriginAllowed(origin)
    },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    clients[ws] = true

    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            delete(clients, ws)
            break
        }
        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <-broadcast
        for client := range clients {
            err := client.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleConnections)
    go handleMessages()

    log.Println("Server starting at :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}