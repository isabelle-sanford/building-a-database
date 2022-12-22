package main

import (
	"binary"
	"bytes"
	"encoding/binary"
	"fmt"
)

// is blocksize in bytes or bits?
const BLOCKSIZE int = 4096 // os.Getpagesize()
const INTSIZE int = 64 / 8 //strconv.IntSize


type Page struct {
	contents []byte
	buffer bytes.Buffer
}

// for log pages (will not be used for now)
func makePage(b []byte) (p Page) {
	bb := bytes.NewBuffer(b)
	p = Page{b, bb}
	return
}

// for data pages
func makePage(blocksize int) (p Page) {
	var b [blocksize]byte
	bb := bytes.NewBuffer(b)
	p = Page{b, bb}
	return
}


func (p Page) getInt(offset int) int {
	var b int

	offsetLoc := p.contents[offset:offset + INTSIZE]

	buf := bytes.NewReader(offsetLoc) // could i take out and just do offsetloc
	err := binary.Read(buf, binary.LittleEndian, b)

	if err != nil {
		fmt.Println("binary.Read failed: ", err)
	}

	return b
}

func (p Page) setInt(offset int, value int) {
	locToWrite := p.contents[offset:offset + INTSIZE]
	writeTo := bytes.NewBuffer(locToWrite)

	err := binary.Write(writeTo, binary.LittleEndian, value)

	if err != nil {
		fmt.Println("binary.Write failed: ", err)
	}
}