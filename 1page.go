package main

import (
	"encoding/binary"
	"fmt"
)

// is blocksize in bytes or bits?
const BLOCKSIZE int64 = 64 //! temporary for ease of testing // os.Getpagesize()
const INTSIZE int = 8 //strconv.IntSize

// slice instead of array for now because write is unhappy >:(
type Page struct {
	contents []byte
	//buffer bytes.Buffer
}

// for data pages
func makePage(blocksize int64) (p Page) {
	var b []byte = make([]byte, blocksize)// temp replacement for blocksize
	//bb := bytes.NewBuffer(b)
	p = Page{b} //, *bb}
	return
}

func (p Page) getInt(offset int) int64 {
	var b int64

	offsetLoc := p.contents[offset:offset + INTSIZE]

	fmt.Printf("Pre-decoding binary: %x\n", offsetLoc)

	b,_ = binary.Varint(offsetLoc)

	// error handling

	return b
}

func (p *Page) setInt(offset int, value int64) {
	inBytes2 := make([]byte, INTSIZE)
	//var inBytes []byte 
	n := binary.PutVarint(inBytes2, value)

	fmt.Printf("Wrote %d bytes\n", n)

	fmt.Printf("Translation is: %d or %v\n", inBytes2[0:3], inBytes2)

	for i := 0; i < int(INTSIZE); i++ {
		p.contents[offset + i] = inBytes2[i]
	}
}

func (p Page) getBytes(offset int) []byte {
	len := p.getInt(offset)

	toret := p.contents[offset + INTSIZE:offset + INTSIZE + int(len)]

	return toret
}

func (p *Page) setBytes(offset int, b []byte) {
	len := cap(b)
	p.setInt(offset, int64(len))

	start := offset + len

	for i := 0; i < len; i++ {
		p.contents[start + i] = b[i]
	}
}

// func main() {
// 	var test int64 = 1029388

// 	var p Page = makePage(BLOCKSIZE)

// 	fmt.Println("Trying to insert the integer ", test)
// 	p.setInt(128, test)

// 	fmt.Printf("The contents now look like %x\n", p.contents[128:200])

// 	retint := p.getInt(128)

// 	//etcint := p.getInt(0)

	
// 	fmt.Println(retint)
// 	//fmt.Println(etcint)

// 	//fmt.Println(p.contents)
// }