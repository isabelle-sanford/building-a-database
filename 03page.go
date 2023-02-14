package main

import (
	"encoding/binary"
	"fmt"
)

// is blocksize in bytes or bits?
const BLOCKSIZE int = 64 //! temporary for ease of testing // os.Getpagesize()
const INTSIZE int = 8    //strconv.IntSize

// slice instead of array for now because write is unhappy >:(
type Page struct {
	contents []byte
}

func (p *Page) String() string {
	// todo make actual Builder later (& maybe improve in general?)

	var s string
	for i := 0; i < len(p.contents); i += 8 {
		s += fmt.Sprint(p.contents[i:i+8], "\n")
	}

	return s
}

// constructor for data pages
func makePage(blocksize int) (p *Page) {
	var b []byte = make([]byte, blocksize)
	p = &Page{b}
	return
}

// page constructor for log pages
func makeLogPage(b []byte) *Page {
	return &Page{b}
}

// get bytes from offset and convert into an int64
func (p *Page) getInt(offset int) int64 {
	offsetLoc := p.contents[offset : offset+INTSIZE]
	b, _ := binary.Varint(offsetLoc)
	// todo error handling
	return b
}

// turn value into bytes array and insert at offset
func (p *Page) setInt(offset int, value int64) {
	inBytes := make([]byte, INTSIZE)
	n := binary.PutVarint(inBytes, value)

	for i := 0; i < int(INTSIZE); i++ {
		if i >= n {
			p.contents[offset+i] = byte(0)
		} else {
			p.contents[offset+i] = inBytes[i]
		}
	}
}

// return bytes array at offset, length determined by int val at offset
func (p *Page) getBytes(offset int) []byte {
	len := p.getInt(offset)
	toret := p.contents[offset+INTSIZE : offset+INTSIZE+int(len)]
	return toret
}

// insert bytes array at offset
func (p *Page) setBytes(offset int, b []byte) {
	len := len(b) // should handle len > blocksize error case
	p.setInt(offset, int64(len))
	start := offset + INTSIZE

	// is there a better way to do this?
	for i := 0; i < len; i++ {
		p.contents[start+i] = b[i]
	}
}

// read and return string from page at offset
func (p *Page) getString(offset int) string {
	b := p.getBytes(offset)
	return string(b)
}

// write val into page at offset
func (p *Page) setString(offset int, val string) {
	b := []byte(val)
	p.setBytes(offset, b)
}

func testPage() {
	var test int64 = 1029388

	var p *Page = makePage(400)

	fmt.Println("Trying to insert the integer ", test)
	p.setInt(128, test)

	fmt.Printf("The contents now look like %x\n", p.contents[128:200])

	retint := p.getInt(128)

	fmt.Println(retint)

	//fmt.Println(p.contents)
	fmt.Print("Page tests complete. \n\n")
}
