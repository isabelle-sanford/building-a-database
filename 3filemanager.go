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

// TODO TEST
func (fm *FileMgr) makeBlock(filename string, blknum int) BlockId {
	// fmt.Printf("Running makeBlock(%s, %d)\n", filename, blknum)
	// fmt.Println(fm.openFiles)
	fm.getFile(filename) // making sure file is actually created
	// (and thus in openFiles)
	bi := BlockId{filename, blknum}
	for blknum >= fm.openFiles[filename] {
		//fmt.Printf("That block does not exist yet!\n")
		fm.appendNewBlock(filename) 
		p := makePage(fm.blocksize)
		fm.writeBlock(bi, p)
		//fmt.Printf("Making (new) block %v, currently containing %v\n", bi, p)
	}

	

	return BlockId{filename, blknum}
}

func makeFileMgr(dbDir string, blocksize int) FileMgr {
	_, err := os.Open(dbDir) // ! no // might need to close? 
	isNew := false
	if errors.Is(err, os.ErrNotExist) { // when do you change from isNew? 
		isNew = true
		os.MkdirAll(dbDir, 0777) // !! perms??
	}
	// remove any leftover temp tables

	openFiles := make(map[string]int)

	return FileMgr{dbDir, isNew, openFiles, blocksize}
}
func (fm *FileMgr) readBlock(blk BlockId, p *Page) (bool) {
	var f *os.File = fm.getFile(blk.filename)

	//var b []byte = make([]byte, BLOCKSIZE)
	_, err := f.ReadAt(p.contents, int64(blk.blknum * fm.blocksize))

	worked := true

	if err != nil {
		fmt.Println("Failed to read block in file: ", blk, err)
		worked = false
	}

	//fmt.Printf("Reading block %v (size %d) and returning %v\n", blk, n, p)

	defer f.Close()
	return worked // could maybe just return bool?
}

// write given page to block
func (fm *FileMgr) writeBlock(blk BlockId, p *Page) bool {
	var f *os.File = fm.getFile(blk.filename)

	_, err := f.WriteAt(p.contents, int64(fm.blocksize * blk.blknum))

	worked := true

	if err != nil {
		fmt.Println("Failed to write block to file: ", err)
		worked = false
	}


	defer f.Close()
	return worked
}

// PRIVATE TO FILE MANAGER
// attach to file manager object? 
// return opened file, create first if it doesn't exist
func (fm *FileMgr) getFile(filename string) *os.File { // might need pointer?
	_, ok := fm.openFiles[filename]

	path := filename // fmt.Sprintf("../%s/%s", fm.dbDir, filename)
	//fmt.Println(path)

	if ok { // filename is in files
		f, err := os.OpenFile(path, os.O_RDWR,0666) // ! PERM STUFF

		if err != nil {
			fmt.Println("Failed to open file: ", err)
		}

		return f
	} else {
		dbTable, err := os.Create(path)
		if err != nil {
			fmt.Println("Failed to create file: ", err)
			return nil
		}

		fm.openFiles[filename] = 0
		return dbTable
	}

}

// new empty block to the end of a file
func (fm *FileMgr) appendNewBlock(filename string) BlockId {

	var f *os.File = fm.getFile(filename)


	newblknum := fm.openFiles[filename] 
	blk := BlockId{filename,newblknum} 

	fm.openFiles[filename]++

	var b []byte = make([]byte, fm.blocksize) // ! using BLOCKSIZE does... something

	f.WriteAt(b, int64(newblknum * fm.blocksize)) 

	f.Close()
	
	return blk
}



func testFileMgr() {
	fm := makeFileMgr("mydb", 64)

	b0 := fm.appendNewBlock("testfile") // b0
	b1 := fm.appendNewBlock("testfile")
	b2 := fm.appendNewBlock("testfile") // b2

	var test int64 = 1029388
	var p *Page = makePage(fm.blocksize)
	p.setInt(16, test)
	p.setString(34, "test")

	var pp *Page = makePage(fm.blocksize)
	pp.setInt(0, 809)
	pp.setString(30, "hello world")

	// ! MUST use distinct pages for reading, else what's written to the page-to-read stays there even when you overwrite it with a different block to read
	// I think
	// not 100% sure
	var p0 *Page = makePage(fm.blocksize)
	// var p1 Page = makePage(fm.blocksize)
	// var p2 Page = makePage(fm.blocksize)
	// var p3 Page = makePage(fm.blocksize)

	fmt.Printf("FM files: %v\n", fm.openFiles)
	fmt.Printf("Test page looks like: %v\n", p.contents)


	fmt.Println("Writing test page to block 0")

	fm.writeBlock(b0, p)
	//

	fm.readBlock(b0, p0)
	fmt.Println("block 0: ", p0)

	fm.readBlock(b1, p0)
	fmt.Println("block 1: ", p0)

	fm.readBlock(b2, p0)
	fmt.Println("block 2: ", p0)

	fm.writeBlock(b2, pp)

	fm.readBlock(b2, p0)
	fmt.Println("block 2 post write: ", p0)

	fmt.Println("FileMgr testing complete")

}