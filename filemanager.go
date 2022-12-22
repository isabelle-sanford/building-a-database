package main // for now

import (
	"os"

	"../page"
)

var dbDir string 

const BLOCKSIZE int = 256 // ! temp, real on this computer is 4096
const INTSIZE int = 8 //strconv.IntSize

var isNew bool = false

var openFiles map[string]int // filename, # blocks

type BlockId struct {
	filename string
	blknum int // index of location within file
}


func readInt(blk BlockId, p page.Page) int {
	var f os.File = getFile(blk.filename)

	n := f.readAt([INTSIZE]byte, blk.blknum * BLOCKSIZE)

	return n
}

func readBytes(blk BlockId, p page.Page) []byte {
	var f os.File = getFile(blk.filename)

	n := f.readAt([INTSIZE]byte, blk.blknum * BLOCKSIZE)

	bb := f.readAt([n]byte, blk.blknum * BLOCKSIZE + INTSIZE)

	return bb // ?
}

func getFile(filename string) os.File { // might need pointer?
	_, ok := openFiles[filename]

	if ok { // filename is in files
		return os.Open(filename)
	} else {
		var dbTable os.File = os.Create(filename)
		openFiles[filename] = 0
		return dbTable
	}

}

func appendNewBlock(filename string) BlockId {

	var f os.File = getFile(filename)


	newblknum := openFiles[filename] 
	openFiles[filename]++

	blk := BlockId{filename,newblknum} 

	var b [BLOCKSIZE]byte

	f.WriteAt(b, BLOCKSIZE * blk.blknum) 
	
	return blk
}

func writeBlock(blk BlockId, p page.Page) {
	f := getFile(blk.filename)

	f.writeAt(p.contents, BLOCKSIZE * blk.blknum)
}

