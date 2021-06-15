package raft

import (
	"fmt"
)

type worker struct {
	MulticastJob chan [3]string // url+action,key,val
	JobC chan interface{}
	exit chan bool
}
func NewWorker() *worker{
	w:= worker{
		MulticastJob: make(chan [3]string),
		JobC: make(chan interface{}),
		exit: make(chan bool),
	}
	go func() {
		for{
			select {
			case j := <-w.JobC:
				sendToLeader(j)
			case j := <-w.MulticastJob:
				multicastChange(j)
			case <- w.exit:
				fmt.Println("Closing worker")
				close(w.exit)
				return
			}
		}
	}()
	return &w
}

func sendToLeader(j interface{}) {

}
func (w *worker)Quit(){
	w.exit <- true
	close(w.JobC)
}
var multicastChange = func(urlArgs [3]string) {
	url,key,val := urlArgs[0],urlArgs[1],urlArgs[2]
	fmt.Println(url,key,val)
}