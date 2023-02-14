package main

import "fmt"

// note: first byte in logPage is integer specifying
// loc of end of last record
type LogMgr struct {
	fm           *FileMgr //
	logfile      string   // filename of log file
	currlsn      int
	currblock    BlockId
	logPage      Page // ? could be *Page? not sure which to use
	lastSavedLSN int
}

// LOG MANAGER----------
// log manager constructor
func makeLogMgr(fm *FileMgr, logfile string) LogMgr {
	logsize := fm.openFiles[logfile]
	currBlock := BlockId{logfile, logsize - 1} // 0?
	lgpage := makePage(fm.blocksize)

	if logsize == 0 {
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

// appends records to a block in BACKWARDS order, i.e. oldest rec is last
func (lm *LogMgr) append(rec []byte) int {
	boundary := int(lm.logPage.getInt(0))
	recsize := len(rec) // + INTSIZE
	needBytes := recsize + INTSIZE

	// individual record is too big to fit on 1 page
	if needBytes > lm.fm.blocksize {
		return -1
	}

	// not enouch space to fit this rec
	if boundary-needBytes < INTSIZE {
		lm.flush()
		lm.currblock = lm.appendNewBlock()
		boundary = int(lm.logPage.getInt(0)) // ? isn't this always blocksize?
	}
	recpos := boundary - needBytes

	lm.logPage.setInt(0, int64(recpos)) // set byte 0 to loc of newest (smallest) record
	lm.logPage.setBytes(recpos, rec)    // actually write record to place

	lm.currlsn += 1
	return lm.currlsn
}

// flush the current log page if someone wants to flush a lsn past
// your last saved
func (lm *LogMgr) flushLSN(lsn int) {
	if lsn >= lm.lastSavedLSN { // if not already saved
		lm.flush()
	}
}

// could also return a hasNext func?
// returns function which returns next record from
// newest log until end, then nil
func (lm *LogMgr) iterator() func() []byte {
	lm.flush() // make sure all records are on disk

	pg := makePage(lm.fm.blocksize)
	lsn := lm.lastSavedLSN // ???
	blknum := lm.currblock.blknum

	lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg)
	recpos := int(pg.getInt(0))

	return func() []byte {
		if recpos >= lm.fm.blocksize {
			blknum -= 1
			if blknum < 0 {
				return nil
			}
			lm.fm.readBlock(BlockId{lm.logfile, blknum}, pg)
			recpos = int(pg.getInt(0))
		}
		lsn-- // not actually returning it or using anywhere?
		ret := pg.getBytes(recpos)

		recpos += int(pg.getInt(recpos)) + INTSIZE

		return ret
	}
}

// aux for logmgr
// append new log block to logfile
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

	printLogRecords := func(msg string) {
		fmt.Println(msg)
		iter := lm.iterator()
		next := iter()

		for {
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
		b := make([]byte, pos+INTSIZE)
		p := makeLogPage(b)
		p.setString(0, s)       // sets s into first bytes
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
