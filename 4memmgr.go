package main

// probs don't need constructor here?
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

type BufferMgr struct {
	fm FileMgr // idk if needed?
	lm LogMgr // idk if needed? 
	numbuffs int 
	bufferpool []Buffer 
	numavailable int
}


// LOG MANAGER----------
func makeLogMgr(fm FileMgr, logfile string, blocksize int) LogMgr {
	logsize := fm.openFiles[logfile]
	currBlock := BlockId{logfile, logsize - 1}
	lgpage := makeLogPage(make([]byte, fm.blocksize))

	if (logsize == 0) {
		currBlock = appendNewBlock() // ! sigh
		fm.readBlock(currBlock, lgpage)
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

// filling in BACKWARDS
func (lm LogMgr) append(rec []byte) int {
	recsize := len(rec) // + INTSIZE
	if recsize > lm.fm.blocksize {
		return -1
	}
	if recsize + lm.currlsn < lm.fm.blocksize {
		lm.logPage.setBytes(lm.currlsn - recsize, rec)
		lm.currlsn -= recsize
		return lm.currlsn
	}
	lm.flush()
	// anything flush doesn't do
	lm.currblock = lm.appendNewBlock()
	lm.logPage.setBytes(0, rec)
	lm.currlsn = recsize 
	return lm.currlsn
}

func (lm LogMgr) flush() {
	lm.currblock.blknum += 1 // ?
	lm.fm.writeBlock(lm.currblock, lm.logPage)
	lm.lastSavedLSN = lm.currlsn
	// probs should try and fit some into prev blocks and all that
}

func (lm LogMgr) iterator() func() []byte {

	pg := makePage(lm.fm.blocksize)
	lsn := lm.fm.blocksize + 1 // ???
	blknum := lm.currblock.blknum
	lenret := 0

	return func() []byte {
		
		if lsn > lm.fm.blocksize {
			blknum -= 1
			pg, _ = lm.fm.readBlock(lm.currblock, pg)
			// todo find first lsn in block
		}
		lsn += lenret //! ???? what val here
		ret := pg.getBytes(lsn)
		lenret = len(ret)
		return ret
		// what happens when run out completely
	}
} // ! should be Iterator<[]byte>

// aux for logmgr

func (lm LogMgr) appendNewBlock() BlockId{
	blk := lm.fm.appendNewBlock(lm.logfile)
	lm.logPage.setInt(0, int64(lm.fm.blocksize))
	lm.fm.writeBlock(blk, lm.logPage)
	return blk
}




// BUFFER MANAGER------------
func (bm BufferMgr) pin(blk BlockId) Buffer {}
func (bm BufferMgr) unpin(buff Buffer) {}
func (bm BufferMgr) available() int {}
func (bm BufferMgr) flushAll(txnum int) {}

type Buffer struct {
	fm FileMgr 
	lm LogMgr
	contents Page 
	blk BlockId 
	pins int 
	txnum int // can you do default vals for structs? (that aren't just 0)
	lsn int 
}

//func (bf Buffer) contents() Page {} // probs unneeded
//func (bf Buffer) block() BlockId {} // probs unneeded
func (bf Buffer) isPinned() bool {} 
func (bf Buffer) setModified(txnum int, lsn int) {}
func (bf Buffer) modifyingTx() int {}