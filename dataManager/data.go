package dataManager

import (
	"fmt"
	"strings"
	"sync"
	"time"
	
	"github.com/emirpasic/gods/trees/btree"
)
const total = 1000
type Database struct {
	lock sync.Mutex
	strTree *btree.Tree
}
func timeEstimate(t time.Time){
	fmt.Println("Time taken : - ", fmt.Sprintf("%dns" ,
		time.Since(t).Microseconds()))
}
func (d *Database) SetData(msg string)bool{
	if key := strings.Split(msg,"="); d.strTree.Size() <= total&& len(key) == 2 {
		d.lock.Lock()
		d.strTree.Put(key[0],key[1])
		d.lock.Unlock()
		return true
	}
	return false
}
func (d *Database) DeleteData(msg string)[]byte{
	d.lock.Lock()
	d.strTree.Remove(msg)
	d.lock.Unlock()
	return []byte("OK")
}
func (d *Database) GetData(msg string)[]byte{
	d.lock.Lock()
	v, ok:= d.strTree.Get(msg)
	d.lock.Unlock()
	if ok {
		return []byte(fmt.Sprintf("%v",v))
	}
	return []byte("Not found")
}
func (d *Database)HandleReq(m []byte)[]byte{
	rsp := []byte("Error")
	
	switch string(m)[0] {
	case '>':
		if d.SetData(string(m)[1:]) {rsp = []byte("Ok")}
	case '<':
		rsp = d.GetData(string(m)[1:])
	case '-':
		rsp = d.DeleteData(string(m)[1:])
	}
	return rsp
}

func Init() *Database{
	d:= Database{
		lock: sync.Mutex{},
		strTree: btree.NewWithStringComparator(10),
	}
	return &d
}