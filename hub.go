package main

import (
	"./pkg"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"html"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan message)           // broadcast channel
//var broadcast = make(chan pkg.ChangeData)
type message struct {
	addr string
	text string
}
// Configure the upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

var dbconn pkg.DB

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("static"))
	err := dbconn.NewDB()
	if err!=nil{
		fmt.Print(err)
	}
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	err = http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r,nil)
	if err != nil {
		log.Fatal(err)
	}
	data,err := dbconn.OnConnection()
//	data,err := dbconn.OnJSONConnection()

	fmt.Print("Connected!\n")
	if err != nil {
		log.Fatal(err)
	}
	err = ws.WriteMessage(websocket.TextMessage,[]byte(data))
	//err = ws.WriteMessage(websocket.TextMessage,data)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// Register our new client
	clients[ws] = true

	for {
		var msg []byte
		// Read in a new message as JSON and map it to a Message object
		_,msg,err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}

		err = dbconn.OnRead(string(msg),ws.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}
		// Send the newly received message to the broadcast channel
		ToChan :=message{
			addr:ws.RemoteAddr().String(),
			text: string(msg),
		}

		broadcast <-ToChan
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte("<p>"+
				html.EscapeString(msg.addr)+
				": "+
				html.EscapeString(msg.text)+"</p>"))
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func handleJSONMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		str,err :=json.Marshal(msg)
		if err != nil {
			log.Printf("error: %v", err)
		}
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, str)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}