package term

import (
	"fmt"
	"strings"
	"time"
	
	raft2 "ishan/dxxxxx/RedissDB/raft"
)
var execCommand = func(cmd string, raft *raft2.Raft)string{
	if cmd == "" {return ""}
	if raft.State == raft2.LEADER {
		raft.InformPeers([]byte(cmd))
		return string(raft.Rediss.HandleReq([]byte(cmd)))
	}else if raft.State == raft2.FOLLOWER{
		if strings.Compare(string(cmd[0]), "<") == 0 {
			return string(raft.Rediss.HandleReq([]byte(cmd)))
		}else {
			raft.InformPeers([]byte(cmd))
		}
	}
	return "Ok"
}
func Start(raft *raft2.Raft){
	time.Sleep(4 * time.Second)
	fmt.Println("Connected to database - ")
	for  {
		var first string
		fmt.Printf("-")
		fmt.Scanln(&first)
		fmt.Println(execCommand(first, raft))
	}
}
