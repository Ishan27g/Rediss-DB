package raft

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	
	ll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/gin-gonic/gin"
)

func (r *Raft) heartBeat(context *gin.Context) {
	if r.State == LEADER {
		r.log("i am leader, cannot receive heartbeat")
		context.JSON(http.StatusForbidden, gin.H{"Error": "i am " +
			"leader/candidate, cannot receive heartbeat"})
		return
	}
	if r.State == FOLLOWER {
		if r.Following{
			r.CommChannels.heartBeatTimeoutChan <- true // end previous timer
			go r.startTimer()
		}else {go r.startTimer()}
	}
}

var getPeers = func() *ll.List{
	list := ll.New()
	total, _ := strconv.Atoi(os.Getenv("PEERS"))
	myId, _ := strconv.Atoi(os.Getenv("INSTANCE"))
	for peer:=1; peer<=total; peer++{
		if myId != peer{
			p := strconv.Itoa(peer)
			o := os.Getenv("PEER"+p)
			n, _ := strconv.Atoi(o[1:])
			n = n + 10
			list.Add("http://localhost:" + strconv.Itoa(n))
		}
	}
	return list
}


func (r *Raft) startTimer() {
	r.Following = true
	randomDelay := time.Duration(electionTimeout + randomInt())
	select {
	case <-time.After(randomDelay * time.Millisecond):
		r.startElection()
		break
	case <-r.CommChannels.heartBeatTimeoutChan :
		break
	}
}

func (r *Raft) UpdateChange(w string) {
	it := r.Peers.Iterator()
	switch r.State {
	case LEADER:
		for it.Next() {
			_, url := it.Index(), it.Value()
			go request(fmt.Sprintf("%v%s", url,w), nil)
		}
	default:
		for it.Next() {
			_, url := it.Index(), it.Value()
			u := fmt.Sprintf("%v", url)
			if strings.Contains(u, r.Leader) { // inform only leader
				request(fmt.Sprintf("%v%s", url,w),nil)
				return
			}
		}
	}
}

func (r *Raft) StartElection(){
	r.State = CANDIDATE
	r.timeoutLeader(0)
}

func (r *Raft) startHeartbeat(){
	for {
		it := r.Peers.Iterator()
		time.Sleep(electionTimeout * time.Millisecond)
		for it.Next() {
			_, url := it.Index(), it.Value()
			go request(fmt.Sprintf("%v", url)+ "/leaderHeartbeat", nil)
		}
	}
}
// set as candidate before starting
func (r *Raft)timeoutLeader(flag int) {
	if r.State != CANDIDATE && flag == 0{
		return
	}
	randomDelay := time.Duration(electionTimeout + randomInt())
	select {
	case <-time.After(randomDelay * time.Millisecond):
		r.startElection()
		break
	case <-r.CommChannels.electionTimeoutChan :
		break
	}
}
func (r *Raft)vote(context *gin.Context) {
	if r.State == LEADER {
		vote := VoteResp{Vote: VoteNo}
		context.JSON(http.StatusOK, vote)
		r.log("i am leader/follower, cannot vote")
		return
	}
	var req TXReq
	err := context.ShouldBindJSON(&req)
	if err != nil {
		return
	}
	if req.Data.TermCount > r.SelfTermCount {
		// accept leader
		r.CommChannels.electionTimeoutChan <- true // end my election if any
		r.Leader = req.Data.Source
		r.LeaderTermCount = req.Data.TermCount
		r.State = FOLLOWER
		vote := VoteResp{Vote: VoteYes}
		context.JSON(http.StatusOK, vote)
		r.log("New leader - ", r.Leader)
		return
	}else if req.Data.TermCount == r.SelfTermCount {
		// tie breaker based on source
		if strings.Compare(req.Data.Source, r.SelfId) > 0{
			r.CommChannels.electionTimeoutChan <- true // end my election if any
			r.Leader = req.Data.Source
			r.LeaderTermCount = req.Data.TermCount
			r.State = FOLLOWER
			vote := VoteResp{Vote: VoteYes}
			context.JSON(http.StatusOK, vote)
			r.log("New leader after tie breaker- ", r.Leader)
			return
		}
	}else {
		vote := VoteResp{Vote: VoteNo}
		context.JSON(http.StatusOK, vote)
		r.log("Cannot vote")
	}
}
func (r *Raft)mod(context *gin.Context) {
	action := context.Query("action")
	rsp := []byte("Error")
	key, value := context.Query("key"),context.Query("value")
	if strings.Compare("set", action)==0{
		if r.Rediss.SetData(key+"="+value) {
			if r.State == LEADER {
				url := "/peerChange?action=set&key="+key+"&value="+value
				r.UpdateChange(url)
			}
			rsp = []byte("Ok")
		}
	}else if strings.Compare("delete", action)==0{
		rsp = r.Rediss.DeleteData(key)
		if r.State == LEADER {
			url := "/peerChange?action=delete&key="+key
			r.UpdateChange(url)
		}
	}
	
	context.JSON(http.StatusOK, gin.H{"Message" : rsp})
}
func (r *Raft)get(context *gin.Context) {
	x := r.Rediss.GetData(context.Query("key"))
	context.JSON(http.StatusOK, gin.H{context.Query("key") : string(x)})
}
func (r *Raft)set(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{context.Query("key") :
			context.Query("value")})
}

func (r *Raft) startElection() bool{
	r.State = CANDIDATE
	// todo -> move term count update to after being elected?
	
	r.SelfTermCount++
	
	it := r.Peers.Iterator()
	votes := 0
	var wg sync.WaitGroup
	for it.Next() {
		wg.Add(1)
		_, url := it.Index(), it.Value()
		go func(url interface{}) {
			defer wg.Done()
			votes += sendVoteReq(url, r.HttpPort, r.SelfTermCount)
		}(url)
	}
	wg.Wait()
	r.log("election ended - ", votes)
	if votes == r.Peers.Size() {
		r.log("Elected leader")
		r.State = LEADER
		go r.startHeartbeat()
		return true
	}else {
		r.log("restarting election")
		return false
	}
}
var sendVoteReq = func(url interface{}, self string, term int) int {
	u := fmt.Sprintf("%v", url) + "/voteReq"
	reqData := TXReq{
		Action: "",
		Data: action{
			Source:    self,
			TermCount: term,
		},
	}
	rsp, status := request(u, reqData)
	if status == http.StatusOK && rsp.Vote == VoteYes {
		return 1
	}
	return -1
}
