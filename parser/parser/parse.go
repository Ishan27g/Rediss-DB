package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	
	tree "github.com/emirpasic/gods/trees/avltree"
)
var chars = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n",
	"o","p","q","r","s","t","u","v","w","x","y","z","0","1","2","3","4","5",
	"6","7","8","9"}


type sFile struct {
	filename string
	fileLoc string
	lineNum []int
	wordIndex []int
}
type Parser struct {
	lock sync.Mutex
	tree *tree.Tree
}

func newWord(charTree *tree.Tree, word string, sf *sFile) bool{
	
	
	
	childTree, _ := charTree.Get(string(word[0]))
	if childTree == nil{
		charTree.Put(word, sf)
		return true
	}else {
		switch c := childTree.(type) {
		case *tree.Tree:
			c.Put(word, sf)
			charTree.Put(word, c)
			return true
		}
	}
	
	return false
}
func Start()*Parser {
	avl1 := tree.NewWithIntComparator()
	for _, char := range chars {
		avl2 := tree.NewWithStringComparator()
		c,_ := strconv.Atoi(char)
		avl1.Put(c, avl2)
	}
	return &Parser{
		lock: sync.Mutex{},
		tree: avl1,
	}
}
func (p *Parser)AddFile(filePath string)bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	
	lines := make(map[int][]string)
	words := make(map[string]*sFile)
	lineNum := 0
	for scanner.Scan() {
		lineWords := strings.Split(scanner.Text(), " ")
		fmt.Println(lineWords, lineNum, len(lineWords))
		if len(lineWords) == 0 {
			continue
		}
		for _, word := range lineWords {
			words[word] = &sFile{
				filename:  filePath,
				fileLoc:   filePath,
				lineNum:   nil,
				wordIndex: nil,
			}
		}
		lines[lineNum] = lineWords
		lineNum++
	}
	for lineNum, lineStr := range lines {
		for i, word := range lineStr {
			wordLoc := words[word]
			// fmt.Println(i,lineNum)
			wordLoc.wordIndex = append(wordLoc.wordIndex, i)
			wordLoc.lineNum = append(wordLoc.lineNum, lineNum)
			words[word] = wordLoc
		}
	}
	for word, sf := range words {
		w, _ := strconv.Atoi(word)
		child, found := p.tree.Get(w)
		if found {
			switch c := child.(type) {
			case *tree.Tree:
				c.Put(word, sf) // todo append to a list instead of single sf
				p.tree.Put(w, c)
			}
		}
	}
	return false
}
func (p *Parser)Find(word string){
	
	w, _ := strconv.Atoi(string(word[0]))
	child, found := p.tree.Get(w)
	if found {
		switch c := child.(type) {
		case *tree.Tree:
			fmt.Println(c.Get(word))
		}
	}
}