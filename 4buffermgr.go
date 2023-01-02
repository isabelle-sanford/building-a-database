package main

type BufferMgr struct {
	fm           FileMgr // idk if needed?
	lm           LogMgr  // idk if needed?
	numbuffs     int
	bufferpool   []Buffer
	numavailable int
}

type Buffer struct {
	fm    FileMgr // ??
	lm    LogMgr  // ??
	blk   BlockId
	pg    Page
	pins  int
	txnum int // can you do default vals for structs? (that aren't just 0)
	lsn   int
}

// BUFFER MANAGER------------
func makeBufferManager(fm FileMgr, lm LogMgr, numbuffs int) BufferMgr {
	pool := make([]Buffer, numbuffs)

	for i := 0; i < len(pool); i++ {
		pool[i] = makeBuffer(fm, lm)
	}

	return BufferMgr{fm, lm, numbuffs, make([]Buffer, numbuffs), numbuffs}
}
func (bm BufferMgr) pin(blk BlockId) Buffer {
	//
	if bm.numavailable == 0 {
		// idk? wait?
	}
	// IF BUFFER ALREADY EXISTS, ADD 1 PIN AND RETURN

	// IF NOT EXISTENT
	// PICK UNPINNED BUFFER B
	var b Buffer = bm.bufferpool[0] // !
	b.blk = blk
	bm.fm.readBlock(blk, b.pg)
	b.pins = 1
	b.txnum = -1 // indicates no modification // right?
	// ? append log record?

	return b
}
func (bm BufferMgr) pin2(blk BlockId) Buffer {
	b := bm.tryToPin(blk)

	if b == nil {
		// go for WAIT
	}

	return b
}
func (bm BufferMgr) unpin(buff Buffer) { // pointers...?
	buff.unpin()

	if !buff.isPinned() {
		bm.numavailable++
		notifyAll()
	}
}

//func (bm BufferMgr) available() int {} // probs unnecessary
func (bm BufferMgr) flushAll(txnum int) {
	for _, buff := range bm.bufferpool {
		if buff.txnum == txnum {
			buff.flush()
		}
	}
}

// aux

func (bm BufferMgr) tryToPin(blk BlockId) Buffer {
	b := bm.findExistingBuffer(blk)
	if b == nil {
		b = bm.chooseUnpinnedBuffer()

		if b == nil {
			return nil
		}
		b.assignToBlock(blk)
	}
	if !b.isPinned() {
		bm.numavailable--
	}

	b.pin()
	return b
}

func (bm BufferMgr) findExistingBuffer(blk BlockId) Buffer {
	// loop through to try and find buffer that already has that blk
	for _, buff := range bm.bufferpool {
		if buff.blk == blk { // ? is this equality comparison functional
			return buff
		}
	}
	return nil
}

func (bm BufferMgr) chooseUnpinnedBuffer() Buffer {
	// pick a buffer that has no pins and return it
	for _, buff := range bm.bufferpool {
		if !buff.isPinned() {
			return buff
		}
	}
	return nil
}

func (bm BufferMgr) notifyAll() {
	// resume waiting threads to fight for buffer
}

// idk whether to attach to buffmgr or not
func makeBuffer(fm FileMgr, lm LogMgr) Buffer {
	bf := Buffer{fm, lm, BlockId{"", 0}, makePage(fm.blocksize), 0, -1, -1}
	return bf
}

// BUFFER---------------
//func (bf Buffer) contents() Page {} // probs unneeded
//func (bf Buffer) block() BlockId {} // probs unneeded
// func (bf Buffer) modifyingTx() int {} // " (just returns txnum)
func (bf Buffer) isPinned() bool {
	if bf.pins > 0 {
		return true
	}
	return false
}
func (bf Buffer) setModified(txnum int, lsn int) {
	bf.txnum = txnum
	if lsn >= 0 {
		bf.lsn = lsn
	}
}

// aux
func (bf Buffer) flush() {
	if bf.txnum >= 0 {
		bf.lm.flushLSN(bf.lsn)
		bf.fm.writeBlock(bf.blk, bf.pg)
		bf.txnum = -1 // why this here and not in assignToBlock ?
	}
}

func (bf Buffer) assignToBlock(blk BlockId) {
	bf.flush()

	bf.blk = blk
	bf.fm.readBlock(blk, bf.pg)
	bf.pins = 0
	// pins etc or no ?
}

// ehh kinda unnecessary
func (bf Buffer) pin() {
	bf.pins++
}
func (bf Buffer) unpin() {
	bf.pins--
}


func main() {
	
}