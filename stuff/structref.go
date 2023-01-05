package main

type BlockId struct {
	filename string
	blknum   int // index of location within file
}

type Page struct {
	contents []byte
}

type FileMgr struct {
	dbDir     string // might need to be pointer or os.File
	isNew     bool
	openFiles map[string]int
	blocksize int
}

type LogMgr struct {
	fm           *FileMgr //
	logfile      string   // filename of log file
	currlsn      int
	currblock    BlockId
	logPage      Page // could be *Page? not sure which to use
	lastSavedLSN int
	//blocksize int // in filemgr
	// lsn list?
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

type BufferMgr struct {
	fm           *FileMgr // idk if needed?
	lm           *LogMgr  // idk if needed?
	numbuffs     int
	bufferpool   []*Buffer
	numavailable int
}

type BufferList struct {
	buffers map[BlockId]*Buffer
	pins    map[BlockId]int // ! YUCK
	bm      *BufferMgr
}

type Transaction struct {
	txnum    int
	bm       *BufferMgr
	fm       *FileMgr
	bufflist BufferList
}

type Schema struct {
	fields map[string]FieldInfo
}

type FieldInfo struct {
	fldtype int
	length  int
}

type Layout struct {
	schema   Schema
	offsets  map[string]int
	slotsize int
}

type RecordPage struct {
	tx     *Transaction
	blk    BlockId
	layout Layout
}

type TableScan struct {
	tx       *Transaction
	tblname  string
	layout   Layout
	rp       *RecordPage
	filename string // probs should be filemanager (when > 1 file)
	currslot int
}

type RID struct { // Record IDentifier
	blknum int
	slot   int
}
