package main

import "go/scanner"

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

type TableMgr struct {
	tx     *Transaction
	tblcat Layout
	fldcat Layout // ??
}

type StatMgr struct {
	tm *TableMgr
	//tx         *Transaction
	tableStats map[string]StatInfo
	numcalls   int
}

type StatInfo struct {
	tblname   string
	numblocks int
	numrecs   int
	distincts map[string]int
}

type MetadataMgr struct {
	tm *TableMgr
	sm *StatMgr
}

type Constant struct {
	ival  int
	sval  string
	isInt bool
}

type Expression struct {
	val     Constant
	fldname string
	isConst bool
}

type Term struct {
	lhs, rhs Expression
}

type Predicate struct {
	terms []Term
}

type Scan interface {
	beforeFirst()
	next() bool
	getInt(fldname string) (int, bool)
	getString(fldname string) (string, bool)
	getVal(fldname string) (Constant, bool)
	hasField(fldname string) bool
	close()
}

// used by TableScan and SelectScan
type UpdateScan interface {
	Scan
	setInt(fldname string, val int)
	setString(fldname string, val string)
	setVal(fldname string, val Constant)
	insert()
	delete()

	getRid() RID
	moveToRid(rid RID)
}

type SelectScan struct {
	scn  Scan
	pred Predicate
}

type ProjectScan struct {
	scn       Scan
	fieldlist []string
}

type ProductScan struct {
	s1 Scan
	s2 Scan
}

type QueryData struct {
	fields []string
	tables []string // collection?
	pred   *Predicate
}

// used for talking about below Data objects
type UpdateData interface{}

type InsertData struct { // tableName, fields, vals
	tblname string
	flds    []string
	vals    []Constant
}

type DeleteData struct { // tableName, pred
	tblname string
	pred    Predicate
}

type ModifyData struct { // tableName, targetField, newValue, pred
	tblname string
	fldname string
	newval  Expression
	pred    Predicate
}

type CreateTableData struct {
	tblname string
	sch     Schema
}

type CreateViewData struct {
	viewname string
	qrdata   QueryData
}

type CreateIndexData struct {
	idxname string
	tblname string
	fldname string
}

type Lexer struct {
	keywords map[string]bool // might need map
	tok      scanner.Scanner
	currTok  rune // maybe unnecessary tbh
}

type Parser struct {
	lex Lexer
}

type Plan interface {
	open() Scan
	blocksAccessed() int // probs unnecessary tbh but eh
	recordsOutput() int
	distinctValues(fldname string) int
	schema() *Schema
}

type TablePlan struct {
	tx      *Transaction
	tblname string
	layout  Layout
	si      *StatInfo
}

type SelectPlan struct {
	p    Plan
	pred Predicate
}

type ProjectPlan struct {
	p   Plan
	sch *Schema
}

type ProductPlan struct {
	p1, p2 Plan
	sch    *Schema
}

type QueryPlanner interface {
	createPlan(data QueryData, tx *Transaction) Plan
}

type UpdatePlanner interface {
	executeInsert(data InsertData, tx *Transaction) int
	executeDelete(data DeleteData, tx *Transaction) int
	executeModify(data ModifyData, tx *Transaction) int
	executeCreateTable(data CreateTableData, tx *Transaction) int
	executeCreateView(data CreateViewData, tx *Transaction) int
	executeCreateIndex(data CreateIndexData, tx *Transaction) int
}

type BasicQueryPlanner struct {
	mdm MetadataMgr
}

type BasicUpdatePlanner struct {
	mdm MetadataMgr
}

type Planner struct {
	qp QueryPlanner
	up UpdatePlanner
}
