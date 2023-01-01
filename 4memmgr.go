package mydb

type LogMgr struct {
	fm FileMgr // ??
	logfile string
}

func (lg LogMgr) append(rec []byte) int {}
func (lg LogMgr) flush(lsn int) {}
func (lg LogMgr) iterator() []byte {} // ! should be Iterator<[]byte>

type BufferMgr struct {
	fm FileMgr // idk if needed?
	lm LogMgr // idk if needed? 
	numbuffs int 
	bufferpool []Buffer 
	numavailable int
}

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