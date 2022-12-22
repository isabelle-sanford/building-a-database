package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// is blocksize in bytes or bits?
const BLOCKSIZE int64 = 4096 // os.Getpagesize()
const INTSIZE int64 = 64 //strconv.IntSize


type Page struct {
	contents [BLOCKSIZE]byte
	//buffer bytes.Buffer
}

// for log pages (will not be used for now)
// func makePage(b []byte) (p Page) {
// 	bb := bytes.NewBuffer(b)
// 	p = Page{b, bb}
// 	return
// }

// for data pages
func makePage(blocksize int64) (p Page) {
	var b [BLOCKSIZE]byte // temp replacement for blocksize
	//bb := bytes.NewBuffer(b)
	p = Page{b} //, *bb}
	return
}


func (p Page) getInt(offset int64) int64 {
	var b int64

	offsetLoc := p.contents[offset:offset + INTSIZE]

	fmt.Printf("Pre-decoding binary: %x\n", offsetLoc)

	b,_ = binary.Varint(offsetLoc)

	// error handling

	return b
}

func (p Page) setInt(offset int64, value int64) {
	locToWrite := &p.contents[offset]
	fmt.Println("We are now examining the array from bit ", offset, " to bit ", offset + INTSIZE)
	writeTo := bytes.NewBuffer(locToWrite)
	fmt.Println("using buffer ", writeTo)

	err := binary.Write(writeTo, binary.LittleEndian, value)

	fmt.Println("The written-to part of the array now looks like ", locToWrite)

	if err != nil {
		fmt.Println("binary.Write failed: ", err)
	}
}

// func (p Page) setInt1(offset int64, value int64) {
// 	locToWrite := &p.contents[offset]
// 	binary.PutVarint(locToWrite, value)

// 	fmt.Printf("%x\n", p.contents[offset:offset + INTSIZE])
// }

// func (p Page) getBytes(offset int64) []byte {
// 	var length int64 = p.getInt(offset)
// 	var thebytes []byte 

// 	// ewww
// 	buf := bytes.NewReader(p.contents[offset + INTSIZE:offset + INTSIZE + length])
// 	err := binary.Read(buf, binary.LittleEndian, thebytes)

// 	if err != nil {
// 		fmt.Println("binary.Read failed: ", err)
// 	}

// 	return thebytes
// }


func main() {
	var test int64 = 1029388


	var p Page = makePage(BLOCKSIZE)

	fmt.Println("Trying to insert the integer ", test)
	p.setInt(128, test)

	fmt.Printf("The inserted part now looks like %x\n", p.contents[128:192])

	retint := p.getInt(128)

	//etcint := p.getInt(0)

	
	fmt.Println(retint)
	//fmt.Println(etcint)

	//fmt.Println(p.contents)
}