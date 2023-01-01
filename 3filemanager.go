package main // for now

import (
	"errors"
	"fmt"
	"os"
)

type FileMgr struct {
	dbDir string // might need to be pointer or os.File
	isNew bool
	openFiles map[string]int
	blocksize int
}

type BlockId struct {
	filename string
	blknum int // index of location within file
}

func makeFileMgr(dbDir string, blocksize int) FileMgr {
	_, err := os.Open(dbDir) // ! no
	isNew := false
	if errors.Is(err, os.ErrNotExist) { // when do you change from isNew? 
		isNew = true
		os.MkdirAll(dbDir, 0666) // !! perms??
	}
	// remove any leftover temp tables

	openFiles := make(map[string]int)

	return FileMgr{dbDir, isNew, openFiles, blocksize}
}

func (fm FileMgr) readBlock(blk BlockId, p Page) Page {
	var f *os.File = fm.getFile(blk.filename)

	//var b []byte = make([]byte, BLOCKSIZE)
	_, err := f.ReadAt(p.contents, int64(blk.blknum * fm.blocksize))

	if err != nil {
		fmt.Println("Failed to read block in file: ", err)
	}

	f.Close()
	return p // ?
}

// write given page to block
func (fm FileMgr) writeBlock(blk BlockId, p Page) {
	var f *os.File = fm.getFile(blk.filename)

	_, err := f.WriteAt(p.contents, int64(fm.blocksize * blk.blknum))

	if err != nil {
		fmt.Println("Failed to write block to file: ", err)
	}

	f.Close()
}

// PRIVATE TO FILE MANAGER
// attach to file manager object? 
// return opened file, create first if it doesn't exist
func (fm FileMgr) getFile(filename string) *os.File { // might need pointer?
	_, ok := fm.openFiles[filename]

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

		fm.openFiles[filename] = 0
		return dbTable
	}

}

// new empty block to the end of a file
func (fm FileMgr) appendNewBlock(filename string) BlockId {

	var f *os.File = fm.getFile(filename)


	newblknum := fm.openFiles[filename] 
	fm.openFiles[filename]++

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