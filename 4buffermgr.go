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
	pg    *Page
	pins  int
	txnum int
	lsn   int
}

// BUFFER MANAGER------------
// buffer manager constructor
func makeBufferManager(fm *FileMgr, lm *LogMgr, numbuffs int) BufferMgr {
	pool := make([]*Buffer, numbuffs)
	for i := 0; i < len(pool); i++ {
		pool[i] = makeBuffer(fm, lm)
	}

	return BufferMgr{fm, lm, numbuffs, pool, numbuffs}
}

// pin a block to the buffer pool (if possible)
func (bm *BufferMgr) pin(blk BlockId) (*Buffer, error) {
	b, err := bm.tryToPin(blk)
	if err != nil {
		fmt.Println("No buffers available currently. Try again later.")
		return nil, err
	}
	return b, nil
}

// unpin a particular buffer page
func (bm *BufferMgr) unpin(buff *Buffer) {
	buff.unpin()

	if !buff.isPinned() {
		bm.numavailable++
		// todo notifyAll()
	}
}

// flush all buffers with transaction number @txnum to disk
func (bm *BufferMgr) flushAll(txnum int) {
	for _, buff := range bm.bufferpool {
		if buff.txnum == txnum {
			buff.flush()
		}
	}
}

// aux functions

// inner func for pin();
func (bm *BufferMgr) tryToPin(blk BlockId) (*Buffer, error) {
	b := bm.findExistingBuffer(blk)

	// ? could put all of this timing stuff inside chooseUnpinned?
	if b == nil { // block not already in a buffer
		b = bm.chooseUnpinnedBuffer()

		if b == nil { // no unpinned buffers exist
			// todo make this wait for notifyAll instead of timing out
			timeout := time.Minute / 2 // 30 seconds
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

// compares blk to buffers in current pool and returns the pointer if it finds a match, else nil
func (bm *BufferMgr) findExistingBuffer(blk BlockId) *Buffer {
	for _, buff := range bm.bufferpool {
		if buff.blk == blk {
			return buff
		}
	}
	return nil
}

// todo update to clock algorithm
// finds an unpinned buffer in the pool and return it (nil if there are none)
func (bm *BufferMgr) chooseUnpinnedBuffer() *Buffer {
	for _, buff := range bm.bufferpool {
		if !buff.isPinned() {
			return buff
		}
	}
	return nil
}

// todo
func (bm *BufferMgr) notifyAll() {
	// resume waiting threads to fight for buffer
}

// BUFFER---------------

// idk whether to attach to buffmgr or not
// constructor for Buffer
func makeBuffer(fm *FileMgr, lm *LogMgr) *Buffer {
	return &Buffer{fm, lm, BlockId{"", 0}, makePage(fm.blocksize), 0, -1, -1}
}

// true if buffer has at least 1 pin
func (bf *Buffer) isPinned() bool {
	return bf.pins > 0
}

// indicate buffer was modified by txnum, record is lsn
func (bf *Buffer) setModified(txnum int, lsn int) {
	bf.txnum = txnum
	if lsn >= 0 {
		bf.lsn = lsn
	}
}

// aux

func (bf *Buffer) flush() {
	if bf.txnum >= 0 { // if block was modified (?)
		bf.lm.flushLSN(bf.lsn)
		worked := bf.fm.writeBlock(bf.blk, bf.pg)
		bf.txnum = -1 // reset txnum (don't necessarily empty page though)
		if !worked {
			fmt.Printf("Failed to flush a buffer with blockID %v and page %v", bf.blk, bf.pg)
		}
	}
}

// assign a buffer to a specific block (will make block if it doesn't exist)
func (bf *Buffer) assignToBlock(blk BlockId) {
	bf.flush()

	bf.blk = blk
	bf.pins = 0

	ok := bf.fm.readBlock(blk, bf.pg)
	if !ok {
		// try creating the block if it doesn't exist, then try the read again
		bf.fm.makeBlock(blk.filename, blk.blknum) // ???
		ok = bf.fm.readBlock(blk, bf.pg)
		if !ok { // no idea what could cause this
			fmt.Printf("Failed to read block %v into buffer %v\n", blk, bf)
		}
	}
}

func (bf *Buffer) pin() {
	bf.pins++
}

// decrement by one pin (will not go down below 0)
func (bf *Buffer) unpin() {
	if bf.pins > 0 {
		bf.pins--
	}
}

func bufferTest() {
	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(&fm, "logfile")
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

	buff0, _ := bm.pin(b0) // !
	p0 := buff0.pg
	n := p0.getInt(20)
	p0.setInt(20, n+1)
	buff0.setModified(1, 0)
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

	fmt.Println("Buffer testing complete\n")

}

func bufferMgrTest() { // buffer manager test

	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(&fm, "logfile")
	bm := makeBufferManager(&fm, &lm, 3)

	var buff [6]*Buffer
	buff[0], _ = bm.pin(fm.makeBlock("bmtest", 0))
	buff[1], _ = bm.pin(fm.makeBlock("bmtest", 1))
	buff[2], _ = bm.pin(fm.makeBlock("bmtest", 2))
	bm.unpin(buff[1])
	buff[1] = nil

	buff[3], _ = bm.pin(fm.makeBlock("bmtest", 0))
	buff[4], _ = bm.pin(fm.makeBlock("bmtest", 1))

	fmt.Printf("Available buffers: %v\n", bm.numavailable)

	fmt.Println("Attempting to pin block 3...")
	_, err := bm.pin(fm.makeBlock("bmtest", 3))
	fmt.Printf("Result: %v", err)

	fmt.Println("Unpinning buff2 and trying again")
	bm.unpin(buff[2])
	buff[2] = nil
	buff[5], _ = bm.pin(fm.makeBlock("bmtest", 3))

	fmt.Println("Final buffer allocation: ")
	for i, bb := range buff {
		if bb != nil {
			fmt.Printf("Buff[%d] pinned to block %v\n", i, *bb)
		}
	}

	fmt.Println("BufferMgr testing complete")

}
