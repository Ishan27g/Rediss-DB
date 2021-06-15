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
	fmt.Println(d.strTree.Size())
	
	if key := strings.Split(msg,"="); d.strTree.Size() <= total&& len(key) == 2 {
		fmt.Println("Adding to database - ",key[0],key[1])
		d.lock.Lock()
		d.strTree.Put(key[0],key[1])
		v, ok:= d.strTree.Get(key[0])
		fmt.Println(msg, v,ok)
		d.lock.Unlock()
		
		fmt.Println(d.strTree)
		return true
	}
	return false
}
func (d *Database) DeleteData(msg string)[]byte{
	d.lock.Lock()
	d.strTree.Remove(msg)
	d.lock.Unlock()
	fmt.Println("Deleting database")
	_, ok:= d.strTree.Get(msg)
	fmt.Println(msg,ok)
	return []byte("OK")
}
func (d *Database) GetData(msg string)[]byte{
	fmt.Println(d.strTree, msg)
	d.lock.Lock()
	v, ok:= d.strTree.Get(msg)
	d.lock.Unlock()
	fmt.Println("Looking up database - ")
	fmt.Println(msg, ok)
	if ok {
		return []byte(fmt.Sprintf("%v",v))
	}
	return []byte("Not found")
}
func (d *Database)HandleReq(m []byte)[]byte{
	defer timeEstimate(time.Now())
	rsp := []byte("Error")
	
	switch string(m)[0] {
	case '>':
		fmt.Println("Add - " , string(m)[1:])
		if d.SetData(string(m)[1:]) {rsp = []byte("Ok")}
	case '<':
		fmt.Println("Get - " , string(m)[1:])
		rsp = d.GetData(string(m)[1:])
	case '-':
		fmt.Println("Remove - " , string(m)[1:])
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