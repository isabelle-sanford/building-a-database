package main

import (
	"fmt"
	"log"
	"time"
)

type BufferMgr struct {
	fm           *FileMgr // idk if needed?
	lm           *LogMgr  // idk if needed?
	numbuffs     int
	bufferpool   []*Buffer
	numavailable int
}

type Buffer struct {
	fm    *FileMgr // ??
	lm    *LogMgr  // ??
	blk   BlockId
	pg    *Page // probably? buffers should DEF always be pointed to
	pins  int
	txnum int // can you do default vals for structs? (that aren't just 0)
	lsn   int
}

// BUFFER MANAGER------------
func makeBufferManager(fm *FileMgr, lm *LogMgr, numbuffs int) *BufferMgr {
	pool := make([]*Buffer, numbuffs)

	for i := 0; i < len(pool); i++ {
		pool[i] = makeBuffer(fm, lm)
	}

	return &BufferMgr{fm, lm, numbuffs, pool, numbuffs}
}

func (bm *BufferMgr) pin(blk BlockId) (*Buffer, error) {
	fmt.Printf("Trying to pin block %v... ", blk)
	b, err := bm.tryToPin(blk)

	if err != nil {
		fmt.Println("No buffers available currently. Try again later.")
		return nil, err
	}

	fmt.Print("Success!\n")

	return b, nil
}
func (bm *BufferMgr) unpin(buff *Buffer) { // pointers...?
	buff.unpin()

	if !buff.isPinned() {
		bm.numavailable++
		//! notifyAll()
	}
}

//func (bm *BufferMgr) available() int {} // probs unnecessary
func (bm *BufferMgr) flushAll(txnum int) {
	for _, buff := range bm.bufferpool { // don't need pointers here right?
		if buff.txnum == txnum {
			buff.flush()
		}
	}
}

// aux

func (bm *BufferMgr) tryToPin(blk BlockId) (*Buffer, error) {
	b := bm.findExistingBuffer(blk)

	if b == nil {
		//fmt.Println("Did not find existing buffer holding block")
		b = bm.chooseUnpinnedBuffer()

		if b == nil {
			timeout := time.Minute / 3
			deadline := time.Now().Add(timeout)
			for tries := 0; time.Now().Before(deadline); tries++ {
				b = bm.chooseUnpinnedBuffer()
				if b != nil {
					return b, nil
				}
				log.Printf("\nAll buffers in use; retrying in 10 seconds...")
				time.Sleep(time.Second * 10) // every 10 seconds
			}
	
			// todo consider panicking here rather than just returning err
			return nil, fmt.Errorf("no buffer was found after %s\n", timeout)
		}
		b.assignToBlock(blk)
	}
	if !b.isPinned() {
		bm.numavailable--
	}
	b.pin()
	return b, nil
}

func (bm *BufferMgr) findExistingBuffer(blk BlockId) *Buffer {
	// loop through to try and find buffer that already has that blk
	for _, buff := range bm.bufferpool {
		if buff.blk == blk { // ? is this equality comparison functional
			return buff
		}
	}
	return nil
}

func (bm *BufferMgr) chooseUnpinnedBuffer() *Buffer {
	// pick a buffer that has no pins and return it
	for _, buff := range bm.bufferpool {
		if !buff.isPinned() {
			return buff
		}
	}
	return nil
}

func (bm *BufferMgr) notifyAll() {
	// resume waiting threads to fight for buffer
}

// BUFFER---------------
//func (bf Buffer) contents() Page {} // probs unneeded
//func (bf Buffer) block() BlockId {} // ""   ""
// func (bf Buffer) modifyingTx() int {} // "" (just returns txnum)

// idk whether to attach to buffmgr or not
func makeBuffer(fm *FileMgr, lm *LogMgr) *Buffer {
	return &Buffer{fm, lm, BlockId{"", 0}, makePage(fm.blocksize), 0, -1, -1}
}

func (bf *Buffer) isPinned() bool {
	return bf.pins > 0
}

func (bf *Buffer) setModified(txnum int, lsn int) {
	bf.txnum = txnum
	if lsn >= 0 {
		bf.lsn = lsn
	}
}

// aux
func (bf *Buffer) flush() {
	if bf.txnum >= 0 {
		bf.lm.flushLSN(bf.lsn)
		worked := bf.fm.writeBlock(bf.blk, bf.pg)
		bf.txnum = -1 
		
		if !worked {
			fmt.Printf("Failed to flush a buffer with blockID %v and page %v", bf.blk, bf.pg)
		}
	}
}

func (bf *Buffer) assignToBlock(blk BlockId) {
	//fmt.Println("Trying to assign buffer to a block. Flushing...")
	bf.flush()

	bf.blk = blk
	ok := bf.fm.readBlock(blk, bf.pg)

	if !ok {
		fmt.Printf("Failed to read block %v into buffer %v\n", blk, bf)
	}

	bf.pins = 0
}

func (bf *Buffer) pin() {
	bf.pins++
}
func (bf *Buffer) unpin() {
	bf.pins--
}


func bufferTest() {
	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(fm, "logfile")
	bm := makeBufferManager(&fm, &lm, 3)

	fmt.Println("managers made...")

	b0 := fm.makeBlock("bufftest", 0)
	b1 := fm.makeBlock("bufftest", 1)
	b2 := fm.makeBlock("bufftest", 2)
	b3 := fm.makeBlock("bufftest", 3)

	fmt.Println("blocks appended...")

	p := makePage(fm.blocksize)
	fm.writeBlock(b0, p)
	fm.writeBlock(b1, p)
	fm.writeBlock(b2, p)
	fm.writeBlock(b3, p)

	fmt.Println("blocks written to...")
	fmt.Printf("Files in filemanager: %v\n\nStarting pins...\n", fm.openFiles)

	buff0,_ := bm.pin(b0) // !
	p0 := buff0.pg
	n := p0.getInt(20)
	p0.setInt(20, n+1)
	buff0.setModified(1,0)
	fmt.Printf("The new value in b1 is %d\n", n+1)
	bm.unpin(buff0)

	// one of these flushes buff0 to disk
	buff1, _ := bm.pin(b1)
	bm.pin(b2)
	bm.pin(b3)

	bm.unpin(buff1)
	buff1, _ = bm.pin(b0)

	p1 := buff1.pg
	p1.setInt(20, 9999)
	buff1.setModified(1, 0)
	bm.unpin(buff1)

	fmt.Println(bm.bufferpool[2])
	fmt.Println(buff0)

}

func bufferMgrTest() { // buffer manager test 

	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(fm, "logfile")
	bm := makeBufferManager(&fm, &lm, 3)

	var buff [6]*Buffer
	buff[0],_ = bm.pin(fm.makeBlock("bmtest", 0))
	buff[1],_ = bm.pin(fm.makeBlock("bmtest", 1))
	buff[2],_ = bm.pin(fm.makeBlock("bmtest", 2))
	bm.unpin(buff[1])
	buff[1] = nil 

	buff[3],_ = bm.pin(fm.makeBlock("bmtest", 0))
	buff[4],_ = bm.pin(fm.makeBlock("bmtest", 1))

	fmt.Printf("Available buffers: %v\n", bm.numavailable)

	fmt.Println("Attempting to pin block 3...")
	_, err := bm.pin(fm.makeBlock("bmtest", 3))
	fmt.Printf("Result: %v", err)

	fmt.Println("Unpinning buff2 and trying again")
	bm.unpin(buff[2])
	buff[2] = nil 
	buff[5],_ = bm.pin(fm.makeBlock("bmtest", 3))

	fmt.Println("Final buffer allocation: ")
	for i, bb := range buff {
		if bb != nil {
			fmt.Printf("Buff[%d] pinned to block %v\n", i, *bb)
		}
	}

}