package main

import (
	"ishan/dxxxxx/RedissDB/parser/parser"
)

func main(){
	p := parser.Start()
	p.AddFile("./sample.txt")
	p.Find("require")
}
