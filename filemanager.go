package main // for now

import (
	"fmt"
	"os"
)

var dbDir string 

// const BLOCKSIZE int64 = 256 // ! temp, real on this computer is 4096
// const INTSIZE int64 = 8 //strconv.IntSize

var isNew bool = false

var openFiles map[string]int64 = make(map[string]int64) // filename, # blocks

type BlockId struct {
	filename string
	blknum int64 // index of location within file
}



func readBlock(blk BlockId, p Page) Page {
	var f *os.File = getFile(blk.filename)

	//var b []byte = make([]byte, BLOCKSIZE)
	_, err := f.ReadAt(p.contents, blk.blknum * BLOCKSIZE)

	if err != nil {
		fmt.Println("Failed to read block in file: ", err)
	}

	f.Close()
	return p // ?
}

// write given page to block
func writeBlock(blk BlockId, p Page) {
	var f *os.File = getFile(blk.filename)

	_, err := f.WriteAt(p.contents, BLOCKSIZE * blk.blknum)

	if err != nil {
		fmt.Println("Failed to write block to file: ", err)
	}

	f.Close()
}

// PRIVATE TO FILE MANAGER
// return opened file, create first if it doesn't exist
func getFile(filename string) *os.File { // might need pointer?
	_, ok := openFiles[filename]

	if ok { // filename is in files
		f, err := os.OpenFile(filename, os.O_RDWR,0666) // ! PERM STUFF

		if err != nil {
			fmt.Println("Failed to open file: ", err)
		}

		return f
	} else {
		dbTable, err := os.Create(filename)
		if err != nil {
			fmt.Println("Failed to create file: ", err)
			return nil
		}

		openFiles[filename] = 0
		return dbTable
	}

}

// new empty block to the end of a file
func appendNewBlock(filename string) BlockId {

	var f *os.File = getFile(filename)


	newblknum := openFiles[filename] 
	openFiles[filename]++

	blk := BlockId{filename,newblknum} 

	var b []byte = make([]byte, BLOCKSIZE)

	f.Write(b) 

	f.Close()
	
	return blk
}



func main() {
	b0 := appendNewBlock("testfile") // b0
	b1 := appendNewBlock("testfile")
	b2 := appendNewBlock("testfile") // b2

	var test int64 = 1029388
	var p Page = makePage(BLOCKSIZE)
	p.setInt(16, test)

	// ! MUST use distinct pages for reading, else what's written to the page-to-read stays there even when you overwrite it with a different block to read
	// I think
	// not 100% sure
	var p0 Page = makePage(BLOCKSIZE)
	var p1 Page = makePage(BLOCKSIZE)
	var p2 Page = makePage(BLOCKSIZE)
	var p3 Page = makePage(BLOCKSIZE)

	fmt.Println(openFiles)

	//fmt.Println(p.contents)

	fmt.Println(b2)

	writeBlock(b0, p)
	writeBlock(b1, p)

	retpage0 := readBlock(b0, p0)
	fmt.Println("block 0: ", retpage0)

	retpage1 := readBlock(b1, p1)
	fmt.Println("block 1: ", retpage1)

	retpage2 := readBlock(b2, p2)
	fmt.Println("block 2: ", retpage2)

	retpage3 := readBlock(b2, p3)
	fmt.Println("block 2: ", retpage3)



}