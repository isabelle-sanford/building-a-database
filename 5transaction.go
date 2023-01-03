package main

import "fmt"

var nextTxNum int = 0
const END_OF_FILE int = -1

type Transaction struct {
	txnum int
	bm *BufferMgr
	fm *FileMgr
	bufflist BufferList
}

type BufferList struct {
	buffers map[BlockId]*Buffer
	pins map[BlockId]int // ! YUCK
	bm *BufferMgr
}

type buffPinCount struct {buff *Buffer; pinCount int}

// BUFFER LIST--------------
func makeBufferList(bm *BufferMgr) *BufferList {
	// idk about capacity being zero here
	return &BufferList{make(map[BlockId]*Buffer,0),make(map[BlockId]int,0), bm}
}
func (bl *BufferList) getBuffer(blk BlockId) *Buffer {
	return bl.buffers[blk]
}
func (bl *BufferList) pin(blk BlockId) {
	buff, _ := bl.bm.pin(blk)
	bl.buffers[blk] = buff // what if already there? 
	// I think that's an overwrite, if page is in different buffer 
	// you want to have the most recent reference 

	bl.pins[blk]++ // ? IS THIS LEGAL BC ZERO VALUE IS 0?
}
func (bl *BufferList) unpin(blk BlockId) {
	buff := bl.buffers[blk] 
	bl.bm.unpin(buff)
	bl.pins[blk]-- // ! hmm
	// if pins does not contain blk, remove it from buffers
}
func (bl *BufferList) unpinAll() {
	for blk, pinct := range bl.pins {
		buff := bl.buffers[blk]

		// sigh
		for i := 0; i < pinct; i++ {
			bl.bm.unpin(buff)
		}
		 // WAIT ANONYMOUS STRUCTS MEAN I COULD- ok later

		 // remove blk from map here? 
	}

	// todo probably have these clear current list rather than allocating new memory
	bl.pins = make(map[BlockId]int,0)
	bl.buffers = make(map[BlockId]*Buffer)
}

// TRANSACTION-----------
func getNextTxNum() int { // NOT A METHOD 
	nextTxNum++
	fmt.Printf("new transaction: %d\n", nextTxNum)
	return nextTxNum	
}
func makeTransaction(fm *FileMgr, lm *LogMgr, bm *BufferMgr) *Transaction {
	return &Transaction{getNextTxNum(), bm, fm, *makeBufferList(bm)}
}

func (tx Transaction) commit() {}
// func (tx Transaction) rollback() {}
// func (tx Transaction) recover() {}

func (tx Transaction) pin(blk BlockId) {
	tx.bufflist.pin(blk)
}
func (tx Transaction) unpin(blk BlockId) {
	tx.bufflist.unpin(blk)
}

func (tx Transaction) getInt(blk BlockId, offset int) int {
	buff := tx.bufflist.getBuffer(blk)
	return int(buff.pg.getInt(offset)) // >:(
		// todo convert int to int64 in Page getInt func
}
func (tx Transaction) getString(blk BlockId, offset int) string {
	buff := tx.bufflist.getBuffer(blk)
	return buff.pg.getString(offset)
}

func (tx Transaction) setInt(blk BlockId, offset int, val int, okToLog bool)  {
	buff := tx.bufflist.getBuffer(blk)
	lsn := -1 
	if okToLog {
		lsn = tx.txnum // ???? idk 
	}
	p := buff.pg 
	p.setInt(offset, int64(val))
	buff.setModified(tx.txnum, lsn)
}
func (tx Transaction) setString(blk BlockId, offset int, val string, okToLog bool) {
	buff := tx.bufflist.getBuffer(blk)
	lsn := -1 
	if okToLog {
		lsn = tx.txnum // ???? idk 
	}
	p := buff.pg 
	p.setString(offset, val)
	buff.setModified(tx.txnum, lsn)
}

//func (tx Transaction) blockSize() int {} = fm.blocksize
//func (tx Transaction) availableBuffs() int {} = bm.numavailable
//func (tx Transaction) size(filename string) int {} // for concurMgr
func (tx Transaction) append(filename string) BlockId {
	// I think?? 
	return tx.fm.appendNewBlock(filename)
} // ?

// txTest
func main() {
	bufferTest()
	bufferMgrTest()

	// vfm := makeFileMgr("mydb", 400)
	// fm := &vfm
	// vlm := makeLogMgr(fm, "log")
	// lm := &vlm
	// vbm := makeBufferManager(fm, lm, 8)
	// bm := &vbm



	// tx1 := makeTransaction(fm, lm, bm)
}








