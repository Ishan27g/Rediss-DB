package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	
	ll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/gin-gonic/gin"
	
	"ishan/dxxxxx/RedissDB/dataManager"
)

const (
	electionTimeout = 4000
)
const (
	LEADER = iota + 1
	CANDIDATE
	FOLLOWER
	VoteYes
	VoteNo
)
type action struct {
	Source string
	TermCount int
}
type TXReq struct {
	Action string `json:"action"`
	Data   action `json:"data"`
}
type VoteResp struct {
	Vote int `json:"vote"`
}
var randomInt = func() int{
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(3000)
}
type Channel struct {
	electionTimeoutChan chan bool
	heartBeatTimeoutChan chan bool
	ChangeWorker        *worker
}
type Raft struct {
	G               *gin.Engine
	State           int
	SelfId          string
	SelfTermCount   int
	Peers           *ll.List
	Leader          string
	LeaderTermCount int
	Rediss          *dataManager.Database
	CommChannels Channel
	HttpPort     string
	Following bool
}
func (r *Raft)log(msg... interface{}){
	fmt.Println(r.SelfId + "-" + fmt.Sprintf("%v", msg))
}
func (r *Raft) getWorkers()*worker {
	return r.CommChannels.ChangeWorker
}
func(r *Raft)LaunchChangeWorker() {
	for {
		select {
		case w := <- r.getWorkers().JobC:
			r.log("New Change - ", w)
			// r.UpdateChange(w)
		}
	}
}
func (r *Raft) setRoutes() {
	// client -> set a value
	// only Leader can use this method
	// r.G.POST("/modify", r.mod) // set?key=&val=
	
	// only Candidates can use this method
	r.G.GET("/set", r.set) // set?key=&val=
	r.G.GET("/get", r.get) // get?key=
	
	// election candidate asks votes from peers
	r.G.POST("/voteReq", r.vote)
	r.G.GET("/leaderHeartbeat", r.heartBeat)
	
	// peers sends on local change
	r.G.GET("/peerChange", r.mod) // with params
}

func (r *Raft) InformPeers(m []byte){
	msg := string(m)
	fmt.Println("Leader informing peers - ", msg)
	switch msg[0] {
	case '>':
		fmt.Println("Inform Peers Add - " , msg[1:])
		if key := strings.Split(msg[1:],"="); len(key) == 2 {
			url := "/peerChange?action=set&key="+key[0]+"&value="+key[1]
			r.UpdateChange(url)
		}
	case '-':
		fmt.Println("Inform Peers Delete - " , msg[1:])
			url := "/peerChange?action=delete&key="+msg[1:]
			r.UpdateChange(url)
	}
}

var request =  func (url string, data interface{}) (*VoteResp, int){
	fmt.Println("Informing Peer - ", url)
	var resp *http.Response
	var err error
	values := data
	jsonData, err := json.Marshal(values)
	if err != nil {
		fmt.Println(err)
		return nil, -1
	}
	if data == nil{
		resp, err = http.Get(url)
	}else {
		resp, err = http.Post(url, "application/json",
			bytes.NewBuffer(jsonData))
	}
	if err != nil {
		fmt.Println(err)
		return nil, -1
	}
	var res VoteResp
	_ = json.NewDecoder(resp.Body).Decode(&res)
	// fmt.Println(res["data"])
	return &res, resp.StatusCode
}

func Init(port int) *Raft {
	// gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	r := Raft{
		G:               g,
		SelfId:          "service"+os.Getenv("INSTANCE"),
		State:           -1,
		SelfTermCount:   0,
		Peers:           getPeers(),
		Leader:          "",
		LeaderTermCount: 0,
		Rediss:          dataManager.Init(),
		CommChannels:   Channel{
			electionTimeoutChan: make(chan bool),
			heartBeatTimeoutChan: make(chan bool),
			ChangeWorker:        NewWorker(),
		},
		HttpPort: strconv.Itoa(port),
		Following: false,
	}
	r.setRoutes()
	go func() {
		_ = r.G.Run(":"+r.HttpPort)
	}()
	return &r
}