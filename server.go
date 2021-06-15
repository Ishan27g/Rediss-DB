package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	
	raft2 "ishan/dxxxxx/RedissDB/raft"
)
var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var serveWs = func(raft *raft2.Raft,
	w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	d := raft.Rediss
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		
		if raft.State == raft2.LEADER {
			raft.InformPeers(p)
			if err := conn.WriteMessage(messageType, d.HandleReq(p)); err != nil {
				log.Println(err)
				return
			}
		}else if raft.State == raft2.FOLLOWER{
			raft.InformPeers(p)
			if err := conn.WriteMessage(messageType, []byte("Ok...")); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
func main(){
	
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	wsAddr := os.Getenv("PEER"+os.Getenv("INSTANCE"))
	port, _ := strconv.Atoi(os.Getenv("PEER"+os.Getenv("INSTANCE"))[1:])
	port = port + 10
	
	raft := raft2.Init(port)
	
	go raft.StartElection()
	
	// web socket for clients to access DB
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(raft, w, r)
	})
	fmt.Println("\nRediss DB \tWS server - ", wsAddr)
	err = http.ListenAndServe(wsAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
