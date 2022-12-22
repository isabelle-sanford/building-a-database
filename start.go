package main

import (
	// "os" // for getting page size
	// "strconv" // for getting int size

	"io"
)

const BLOCKSIZE int = 4096 // os.Getpagesize()
const INTSIZE int = 64 / 8 //strconv.IntSize

type BlockId struct {
	filename string 
	id int
}
// don't need? : equals, tostring

type Page struct {
	bb []byte // size is BLOCKSIZE if data page
	cc io.ReadWriteSeeker
}

// how to make blksize read as const? 
func makePage(blksize int) (p Page) {
	var b [BLOCKSIZE]byte // temp
	c := bufio.newReadWriter(b) // seeking? 

	p = Page{b,c}
	return
}


func (p Page) getInt1(offset int) int {
	p.cc.Seek(offset, 0)
	var ret [INTSIZE]byte 

	r := p.cc.Read(ret)
	return r
}

func (p Page) setInt1(offset int, value int) {
	p.cc.Seek(offset, 0)
	p.cc.Write(value)
	// error handling
}


func (p Page) getBytes(offset int) {
	p.cc.Seek(offset, 0)
	var bsize int = p.cc.ReadByte()

	p.cc.Read
}