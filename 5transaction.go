package main

import "fmt"

var nextTxNum int = 0

const END_OF_FILE int = -1

type Transaction struct {
	txnum    int
	bm       *BufferMgr
	fm       *FileMgr
	bufflist BufferList
}

type BufferList struct {
	buffers map[BlockId]*Buffer
	pins    map[BlockId]int // ! YUCK
	bm      *BufferMgr
}

type buffPinCount struct {
	buff     *Buffer
	pinCount int
}

// BUFFER LIST--------------
func makeBufferList(bm *BufferMgr) *BufferList {
	// idk about capacity being zero here
	return &BufferList{make(map[BlockId]*Buffer, 0), make(map[BlockId]int, 0), bm}
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
	bl.pins = make(map[BlockId]int, 0)
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

// dummies for now
func (tx *Transaction) commit() {
	tx.bm.flushAll(tx.txnum)
	//tx.lm.flushAll()
}
func (tx *Transaction) rollback() {} // does nothing
func (tx *Transaction) recover()  {}

func (tx *Transaction) pin(blk BlockId) {
	tx.bufflist.pin(blk)
}
func (tx *Transaction) unpin(blk BlockId) {
	tx.bufflist.unpin(blk)
}

func (tx *Transaction) getInt(blk BlockId, offset int) int {
	buff := tx.bufflist.getBuffer(blk)
	return int(buff.pg.getInt(offset)) // >:(
	// todo convert int to int64 in Page getInt func
}
func (tx *Transaction) getString(blk BlockId, offset int) string {
	buff := tx.bufflist.getBuffer(blk)
	return buff.pg.getString(offset)
}

func (tx *Transaction) setInt(blk BlockId, offset int, val int, okToLog bool) {
	buff := tx.bufflist.getBuffer(blk)
	lsn := -1
	if okToLog {
		lsn = tx.txnum // ???? idk
	}
	p := buff.pg
	p.setInt(offset, int64(val))
	buff.setModified(tx.txnum, lsn)
}
func (tx *Transaction) setString(blk BlockId, offset int, val string, okToLog bool) {
	buff := tx.bufflist.getBuffer(blk)
	lsn := -1
	if okToLog {
		lsn = tx.txnum // ???? idk
	}
	p := buff.pg
	p.setString(offset, val)
	buff.setModified(tx.txnum, lsn)
}

// func (tx Transaction) blockSize() int {} = fm.blocksize
// func (tx Transaction) availableBuffs() int {} = bm.numavailable
// func (tx Transaction) size(filename string) int {} // for concurMgr
func (tx *Transaction) append(filename string) BlockId {
	// I think??
	return tx.fm.appendNewBlock(filename)
} // ?

// txTest
func txTest() {
	// setup (need to simplify...)
	vfm := makeFileMgr("mydb", 400)
	fm := &vfm
	vlm := makeLogMgr(fm, "log")
	lm := &vlm
	vbm := makeBufferManager(fm, lm, 8)
	bm := &vbm

	p := makePage(fm.blocksize)

	tx1 := makeTransaction(fm, lm, bm)
	blk := fm.makeBlock("testtx", 1) // does text index blocks from 1?
	tx1.pin(blk)

	tx1.setInt(blk, 80, 1, false)
	tx1.setString(blk, 40, "one", false)
	tx1.commit()

	//fmt.Printf("Buffers after tx1: %v\n", *bm.bufferpool[0])

	tx2 := makeTransaction(fm, lm, bm)
	tx2.pin(blk)
	ival := tx2.getInt(blk, 80)
	sval := tx2.getString(blk, 40)
	fmt.Printf("Initial value at location 80 = %d\nInitial value at location 40 = %s\n", ival, sval)
	newival := ival + 1   // 2
	newsval := sval + "!" // one!
	tx2.setInt(blk, 80, newival, true)
	tx2.setString(blk, 40, newsval, true)

	ival = tx2.getInt(blk, 80)
	sval = tx2.getString(blk, 40)
	fmt.Printf("after set/before commit at location 80 = %d\nat location 40 = %s\n", ival, sval)
	tx2.commit() // flushes (but changes don't wait til here to propagate)

	//fmt.Printf("Buffers after tx2: %v\n", *bm.bufferpool[0])
	fmt.Printf("tx2 after completion: %v\n", tx2)

	tx3 := makeTransaction(fm, lm, bm)
	tx3.pin(blk)
	fm.readBlock(blk, p)
	fmt.Printf("block start of tx3 %v\n", p)
	iival := tx3.getInt(blk, 80)
	ssval := tx3.getString(blk, 40)
	fmt.Printf(
		"start tx3 value at location 80 = %d\ntx3 value at location 40 = %s\n", iival, ssval)
	tx3.setInt(blk, 80, 9999, true)
	fmt.Printf("pre-rollback value at loc 80: %v\n", tx3.getInt(blk, 80))
	tx3.rollback() // does not work

	//fmt.Printf("buffers after tx3: %v\n", *bm.bufferpool[0])

	tx4 := makeTransaction(fm, lm, bm)
	tx4.pin(blk)
	fmt.Printf("post-rollback at location 80 = %d\n", tx4.getInt(blk, 80))
	tx4.commit()

	//fmt.Printf("Buffers after tx4: %v\n", *bm.bufferpool[0])
}
