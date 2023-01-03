package main

import "fmt"

// note: first byte in logPage is integer specifying
// loc of end of last record
type LogMgr struct {
	fm *FileMgr // 
	logfile string // filename of log file
	currlsn int
	currblock BlockId
	logPage Page // could be *Page? not sure which to use
	lastSavedLSN int
	//blocksize int // in filemgr
	// lsn list? 
}



// LOG MANAGER----------
// could just get blocksize from fm? 
func makeLogMgr(fm *FileMgr, logfile string) LogMgr {
	logsize := fm.openFiles[logfile]
	currBlock := BlockId{logfile, logsize - 1} // 0?
	lgpage := makePage(fm.blocksize)

	if (logsize == 0) {
		currBlock = fm.appendNewBlock(logfile)
		lgpage.setInt(0, int64(fm.blocksize))
		fm.writeBlock(currBlock, lgpage)
	}

	return LogMgr{
		fm, 
		logfile, 
		0, 
		currBlock, 
		*lgpage, // I guess? 
		0,
	}
}

// appends to a block in BACKWARDS order, i.e. oldest rec is last
func (lm *LogMgr) append(rec []byte) int {
	boundary := int(lm.logPage.getInt(0))
	recsize := len(rec) // + INTSIZE
	needBytes := recsize + INTSIZE

	// individual log is too big
	if needBytes > lm.fm.blocksize {
		return -1
	}

	// not enouch space to fit this rec
	if boundary - needBytes < INTSIZE {
		lm.flush() // ... was it just flush() before...?
		lm.currblock = lm.appendNewBlock() 
		boundary = int(lm.logPage.getInt(0)) // ? isn't this always blocksize?
	}
	recpos := boundary - needBytes

	lm.logPage.setInt(0, int64(recpos)) // set byte 0 to loc of newest (smallest) record
	lm.logPage.setBytes(recpos, rec) // actually write record to place
	
	lm.currlsn += 1
		
	return lm.currlsn
}

func (lm *LogMgr) flushLSN(lsn int) {
	if lsn >= lm.lastSavedLSN { // if not already saved
		lm.flush()
	}
}

// could also return a hasNext func? 
func (lm *LogMgr) iterator() func() []byte {
	lm.flush()

	pg := makePage(lm.fm.blocksize)
	lsn := lm.lastSavedLSN // ???
	blknum := lm.currblock.blknum

	lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg)
	recpos := int(pg.getInt(0)) 

	fmt.Printf("LSN %d, blknum %d, recpos %d\n", lsn, blknum, recpos)

	return func() []byte {
		
		if recpos >= lm.fm.blocksize  { 
			blknum -= 1
			if blknum < 0 {
				return nil // apparently this is ok? 
			}
			lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg) // probs don't need pg, _ ? 
			recpos = int(pg.getInt(0))
		}
		lsn-- // ehhh // also not actually returning it anywhere? 
		ret := pg.getBytes(recpos)

		recpos += int(pg.getInt(recpos)) + INTSIZE // ! text is different but I THINK this also works

		return ret
	}
} 

// aux for logmgr
func (lm *LogMgr) appendNewBlock() BlockId {
	blk := lm.fm.appendNewBlock(lm.logfile) // !
	lm.logPage.setInt(0, int64(lm.fm.blocksize))
	lm.fm.writeBlock(blk, &lm.logPage) // ? 
	return blk
}

func (lm *LogMgr) flush() {
	lm.fm.writeBlock(lm.currblock, &lm.logPage)
	lm.lastSavedLSN = lm.currlsn
}



func testLogMgr() {
	fm := makeFileMgr("mydb", 80)
	lm := makeLogMgr(&fm, "logfile")

	printLogRecords := func (msg string) {
		fmt.Println(msg)
		iter := lm.iterator()
		next := iter()

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


	fmt.Println("LogMgr testing complete")
}



