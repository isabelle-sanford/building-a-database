package main

import (
	"encoding/binary"
	"fmt"
)

// is blocksize in bytes or bits?
const BLOCKSIZE int = 64 //! temporary for ease of testing // os.Getpagesize()
const INTSIZE int = 8 //strconv.IntSize

// slice instead of array for now because write is unhappy >:(
type Page struct {
	contents []byte
	//buffer bytes.Buffer
}

// for data pages
func makePage(blocksize int) (p Page) {
	var b []byte = make([]byte, blocksize)// temp replacement for blocksize
	//bb := bytes.NewBuffer(b)
	p = Page{b} //, *bb}
	return
}

func makeLogPage(b []byte) Page {
	return Page{b}
}

func (p Page) getInt(offset int) int64 {
	var b int64

	offsetLoc := p.contents[offset:offset + INTSIZE]

	//fmt.Printf("Pre-decoding binary: %x\n", offsetLoc)

	b,_ = binary.Varint(offsetLoc)

	// error handling

	return b
}

func (p *Page) setInt(offset int, value int64) {
	inBytes2 := make([]byte, INTSIZE)
	//var inBytes []byte 
	n := binary.PutVarint(inBytes2, value)


	//fmt.Printf("Translation is: %d or %v\n", inBytes2[0:3], inBytes2)

	for i := 0; i < int(INTSIZE); i++ {
		if i >= n {
			p.contents[offset + i] = byte(0)
		} else {
			p.contents[offset + i] = inBytes2[i]
		}
	}

	//fmt.Printf("Wrote to %d as int %d (%d non-zero bytes i.e. %v)\n", offset, value, n, inBytes2)
}

func (p Page) getBytes(offset int) []byte {
	//fmt.Printf("Trying to get bytes at %d\n", offset)
	len := p.getInt(offset)

	toret := p.contents[offset + INTSIZE:offset + INTSIZE + int(len)]

	return toret
}

func (p *Page) setBytes(offset int, b []byte) {
	//fmt.Printf("Inserting bytes %v...\n", b)
	
	len := len(b)

	//fmt.Printf("Byte conversion of string (cap %d) is %v\n", len, b)

	p.setInt(offset, int64(len))

	start := offset + INTSIZE //+ len

	for i := 0; i < len; i++ {
		p.contents[start + i] = b[i]
	}

	//fmt.Printf("at %d, contents %v\n",start, p.contents)
}

func (p Page) getString(offset int) string {
	b := p.getBytes(offset)
	return string(b)
}

func (p *Page) setString(offset int, val string) {
	

	b := []byte(val)

	//fmt.Printf("String %s converted to bytes %v\n", val, b)

	//fmt.Printf("Trying to set into bytes (%d) at offset %d: %s\n", len(b), offset, val)

	p.setBytes(offset, b)
}


func testPage() {
	var test int64 = 1029388

	var p Page = makePage(BLOCKSIZE)

	fmt.Println("Trying to insert the integer ", test)
	p.setInt(128, test)

	fmt.Printf("The contents now look like %x\n", p.contents[128:200])

	retint := p.getInt(128)

	//etcint := p.getInt(0)

	
	fmt.Println(retint)
	//fmt.Println(etcint)

	//fmt.Println(p.contents)
}