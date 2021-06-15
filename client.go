package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"
	
	"github.com/gorilla/websocket"
)
func worker(url string, job <- chan string, rsp chan <- string){
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	j := <- job
	// fmt.Println("sending: ", j)
	err = c.WriteMessage(websocket.TextMessage, []byte(j))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return
	}
	// log.Printf("received: %s", message)
	rsp <- string(message)
	err = c.WriteMessage(websocket.CloseMessage,
	websocket.FormatCloseMessage(websocket.CloseNormalClosure,
		""))
	if err != nil {
		return
	}
}
func test(rsp chan<- string, ix int){ //nolint:typecheck
	
	jobs := make(chan string, cap(rsp))

	for i := 1; i <= 100; i++ {
		jobs <- ">"+strconv.Itoa(i+(ix*100))+"="+strconv.Itoa(i*100)
		u := url.URL{Scheme: "ws", Host: "localhost"}
		worker(u.String()+":9002/ws",jobs, rsp)
	}
	for i := 1; i <= 25; i++ {
		jobs <- "-"+strconv.Itoa(i+(ix*100))
		u := url.URL{Scheme: "ws", Host: "localhost"}
		worker(u.String()+":9003/ws",jobs, rsp)
	}
	for i := 26; i <= 75; i++ {
		jobs <- "<"+strconv.Itoa(i+(ix*100))
		u := url.URL{Scheme: "ws", Host: "localhost"}
		worker(u.String()+":9001/ws",jobs, rsp)
	}
	
	for i := 75; i <= 100; i++ {
		jobs <- "-"+strconv.Itoa(i+(ix*100))
		u := url.URL{Scheme: "ws", Host: "localhost"}
		worker(u.String()+":9003/ws",jobs, rsp)
	}
	
}
func timeEstimate(t time.Time){
	fmt.Println("Time taken : - ", fmt.Sprintf("%dns" ,
		time.Since(t).Microseconds()))
}
func main(){
	defer timeEstimate(time.Now())
	rsp := make(chan string,200)
	
	var multiplier = 10
	var jobs = 200
	
	for i := 1; i <= multiplier; i++ {
		go test(rsp,i)
	}
	var i = 0
	for i = 0; i < multiplier*jobs; i++ {
		fmt.Println(<- rsp)
	}
	close(rsp)
	fmt.Println(i, multiplier*jobs)
}
