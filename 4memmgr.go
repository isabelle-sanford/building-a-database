package main

import "fmt"

// note: first byte in logPage is integer specifying
// loc of end of last record
type LogMgr struct {
	fm FileMgr // 
	logfile string // filename of log file
	currlsn int
	currblock BlockId
	logPage Page
	lastSavedLSN int
	//blocksize int // in filemgr
	// lsn list? 
}



// LOG MANAGER----------
// could just get blocksize from fm? 
func makeLogMgr(fm FileMgr, logfile string) LogMgr {
	logsize := fm.openFiles[logfile]
	currBlock := BlockId{logfile, logsize - 1} // 0?
	lgpage := makePage(fm.blocksize)

	if (logsize == 0) {
		currBlock = fm.appendNewBlock(logfile)
		lgpage.setInt(0, int64(fm.blocksize))
		fm.writeBlock(currBlock, lgpage)


		// fm.readBlock(currBlock, lgpage) // probs dont need
		// fmt.Printf("lgpage at lgmgr creation: %v\n", lgpage)
	}


	return LogMgr{
		fm, 
		logfile, 
		0, 
		currBlock, 
		lgpage, // I guess?
		0,
	}
}
// appends to a block in BACKWARDS order, i.e. oldest rec is last
func (lm *LogMgr) append(rec []byte) int {


	boundary := int(lm.logPage.getInt(0))
	recsize := len(rec) // + INTSIZE
	needBytes := recsize + INTSIZE

	//fmt.Printf("Boundary %d, Recsize %d, needBytes %d\n", boundary, recsize, needBytes)

	// individual log is too big
	if needBytes > lm.fm.blocksize {
		return -1
	}

	// not enouch space to fit this rec
	if boundary - needBytes < INTSIZE {
		//lm.PRINTBLOCK0("\npre flush in append: ")
		lm.flush() // ... was it just flush() before...?
		// flush func
		// fmt.Printf("\n~~FLUSHING %v~~", lm.currblock)
		// lm.fm.writeBlock(lm.currblock, lm.logPage)
		// lm.lastSavedLSN = lm.currlsn
		
		//lm.PRINTBLOCK0("Just after flush, in append: ")

		lm.currblock = lm.appendNewBlock() 
		
		//fmt.Printf("Switching to next block\n")
		
		//lm.PRINTBLOCK0("After appending new block (in append)")

		boundary = int(lm.logPage.getInt(0)) // ? isn't this always blocksize?

		
	}
	recpos := boundary - needBytes

	lm.logPage.setInt(0, int64(recpos)) // ??
	lm.logPage.setBytes(recpos, rec)
	
	lm.currlsn += 1

	//fmt.Printf("Added %d bytes at offset %d\n", needBytes, recpos)
		
	return lm.currlsn
}
func (lm LogMgr) flushLSN(lsn int) {
	if lsn >= lm.lastSavedLSN { // if not already saved
		lm.flush()
	}
}
// maybe also return a hasNext func? 
func (lm LogMgr) iterator() func() []byte {
	lm.flush()

	pg := makePage(lm.fm.blocksize)
	lsn := lm.lastSavedLSN // ???
	blknum := lm.currblock.blknum


	lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg)
	recpos := int(pg.getInt(0)) 

	fmt.Printf("LSN %d, blknum %d, recpos %d\n", lsn, blknum, recpos)
	//fmt.Printf("Starting with blk %d contents: %v\n", blknum, pg)

	return func() []byte {
		
		if recpos >= lm.fm.blocksize  { 
			// todo check if this is 1st block
			//fmt.Println("Switching block")
			blknum -= 1
			if blknum < 0 {
				return nil // apparently this is ok? 
			}
			lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg) // probs don't need pg, _ ? 
			recpos = int(pg.getInt(0))
			//fmt.Printf("Switching to blk %d contents: %v\n", blknum, pg)
		}
		lsn-- // ehhh // also not actually returning it anywhere? 
		ret := pg.getBytes(recpos)

		//fmt.Printf("Returning recpos %d with lsn %d\n", recpos, lsn)

		recpos += int(pg.getInt(recpos)) + INTSIZE // ! text is different but I THINK this also works

		//fmt.Printf("Shifting to %d\n", recpos)

		return ret
	}
} 

// aux for logmgr
func (lm LogMgr) appendNewBlock() BlockId {
	
	blk := lm.fm.appendNewBlock(lm.logfile) // !
	//fmt.Printf("\nBlock being appended: %v", blk)
	//lm.PRINTBLOCK0("In appendNewBlock right after fm.appendNB:")
	
	lm.logPage.setInt(0, int64(lm.fm.blocksize))
	lm.fm.writeBlock(blk, lm.logPage) // ? 
	return blk
}
func (lm *LogMgr) flush() {
	//fmt.Printf("\n~~FLUSHING %v~~", lm.currblock)
	
	lm.fm.writeBlock(lm.currblock, lm.logPage)
	lm.lastSavedLSN = lm.currlsn

	//lm.PRINTBLOCK0("INSIDE flush after write: ")
	//fmt.Print(lm.logPage)
}




func main() {
	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(fm, "logfile")

	printLogRecords := func (msg string) {
		fmt.Println(msg)
		iter := lm.iterator()
		next := iter()

		//fmt.Printf("next #1: %v", next)
		
		for  {
			if next == nil {
				break
			}
			p := makeLogPage(next)
			s := p.getString(0)
			val := p.getInt(len(s) + INTSIZE)
			fmt.Printf("[%s, %d] ", s, val)
			next = iter()
		}
	}

	// just byte equivalent of s + n
	// seems like you could do this easier ngl
	createLogRecord := func(s string, n int) []byte {
		pos := len(s) + INTSIZE
		b := make([]byte, pos + INTSIZE) 
		p := makeLogPage(b) 
		p.setString(0, s) // sets s into first bytes
		p.setInt(pos, int64(n)) // adds int n to end of s
		return b
	}

	createRecords := func(start int, end int) {
		fmt.Print("\nCreating records: \n")
		for i := start; i <= end; i++ {
			rec := createLogRecord(fmt.Sprintf("record%d", i), i+100)
			lsn := lm.append(rec)
			fmt.Printf("%d ", lsn)
		}
	}



	createRecords(1, 10)
	printLogRecords("\n\nThe log file now has these records---------------- \n")
	createRecords(36, 70)
	lm.flushLSN(65)
	printLogRecords("\nThe log file now has these records: ")


}




// type BufferMgr struct {
// 	fm FileMgr // idk if needed?
// 	lm LogMgr // idk if needed? 
// 	numbuffs int 
// 	bufferpool []Buffer 
// 	numavailable int
// }

// type Buffer struct {
// 	fm FileMgr // ??
// 	lm LogMgr // ??
// 	blk BlockId 
// 	pg Page 
// 	pins int 
// 	txnum int // can you do default vals for structs? (that aren't just 0)
// 	lsn int 
// }

// // BUFFER MANAGER------------
// func makeBufferManager(fm FileMgr, lm LogMgr, numbuffs int) BufferMgr {
// 	pool := make([]Buffer, numbuffs)

// 	for i := 0; i < len(pool); i++ {
// 		pool[i] = makeBuffer(fm, lm)
// 	}

// 	return BufferMgr{fm, lm, numbuffs, make([]Buffer, numbuffs), numbuffs}
// }
// func (bm BufferMgr) pin(blk BlockId) Buffer {
// 	// 
// 	if bm.numavailable == 0 {
// 		// idk? wait? 
// 	}
// 	// IF BUFFER ALREADY EXISTS, ADD 1 PIN AND RETURN

// 	// IF NOT EXISTENT
// 	// PICK UNPINNED BUFFER B
// 	var b Buffer = bm.bufferpool[0] // !
// 	b.blk = blk 
// 	bm.fm.readBlock(blk, b.pg)
// 	b.pins = 1
// 	b.txnum = -1 // indicates no modification // right?
// 	// ? append log record? 

// 	return b
// }
// func (bm BufferMgr) pin2(blk BlockId) Buffer {
// 	b := bm.tryToPin(blk)

// 	if b == nil {
// 		// go for WAIT
// 	}

// 	return b
// }
// func (bm BufferMgr) unpin(buff Buffer) { // pointers...?
// 	buff.unpin()

// 	if !buff.isPinned() {
// 		bm.numavailable++
// 		notifyAll()
// 	}
// }
// //func (bm BufferMgr) available() int {} // probs unnecessary
// func (bm BufferMgr) flushAll(txnum int) {
// 	for _, buff := range bm.bufferpool {
// 		if buff.txnum == txnum {
// 			buff.flush()
// 		}
// 	}
// }

// // aux 

// func (bm BufferMgr) tryToPin(blk BlockId) Buffer {
// 	b := bm.findExistingBuffer(blk)
// 	if b == nil {
// 		b = bm.chooseUnpinnedBuffer()

// 		if b == nil {
// 			return nil
// 		}
// 		b.assignToBlock(blk) 
// 	}
// 	if !b.isPinned() {
// 		bm.numavailable--
// 	}

// 	b.pin()
// 	return b
// }

// func (bm BufferMgr) findExistingBuffer(blk BlockId) Buffer {
// 	// loop through to try and find buffer that already has that blk
// 	for _, buff := range bm.bufferpool {
// 		if buff.blk == blk { // ? is this equality comparison functional
// 			return buff
// 		}
// 	}
// 	return nil
// }

// func (bm BufferMgr) chooseUnpinnedBuffer() Buffer {
// 	// pick a buffer that has no pins and return it 
// 	for _, buff := range bm.bufferpool {
// 		if !buff.isPinned() {
// 			return buff
// 		}
// 	}
// 	return nil
// }

// func (bm BufferMgr) notifyAll() {
// 	// resume waiting threads to fight for buffer
// }



// // idk whether to attach to buffmgr or not
// func makeBuffer(fm FileMgr, lm LogMgr) Buffer {
// 	bf := Buffer{fm, lm, BlockId{"",0}, makePage(fm.blocksize),0,-1, -1}
// 	return bf
// }

// // BUFFER---------------
// //func (bf Buffer) contents() Page {} // probs unneeded
// //func (bf Buffer) block() BlockId {} // probs unneeded
// // func (bf Buffer) modifyingTx() int {} // " (just returns txnum)
// func (bf Buffer) isPinned() bool {
// 	if bf.pins > 0 {return true}
// 	return false
// } 
// func (bf Buffer) setModified(txnum int, lsn int) {
// 	bf.txnum = txnum
// 	if lsn >= 0 {
// 		bf.lsn = lsn
// 	}
// }


// // aux 
// func (bf Buffer) flush() {
// 	if bf.txnum >= 0 {
// 		bf.lm.flushLSN(bf.lsn)
// 		bf.fm.writeBlock(bf.blk, bf.pg)
// 		bf.txnum = -1 // why this here and not in assignToBlock ?
// 	}
// }

// func (bf Buffer) assignToBlock(blk BlockId) {
// 	bf.flush()

// 	bf.blk = blk 
// 	bf.fm.readBlock(blk, bf.pg)
// 	bf.pins = 0
// 	// pins etc or no ? 
// }

// // ehh kinda unnecessary
// func (bf Buffer) pin() {
// 	bf.pins++
// }
// func (bf Buffer) unpin() {
// 	bf.pins--
// }
