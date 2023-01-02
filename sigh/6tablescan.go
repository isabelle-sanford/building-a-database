package mydb

type TableScan struct {
	tx      Transaction
	tblname string
	layout  Layout
	rp       RecordPage
	filename string // probs should be filemanager (when > 1 file)
	currslot int
}

type RID struct { // Record IDentifier
	blknum int
	slot   int
}

func makeTableScan(tx Transaction, tblname string, layout Layout) {
	// MAKE file (bc files are homogeneous)
	filename := tblname + ".tbl"
	rp := makeRecordPage(tx, BlockId{filename, 0}, layout)

	t := TableScan{tx, tblname, layout, rp, filename, 0}

	t.moveToBlock(0)
}

func (t TableScan) close() {
	t.tx.unpin(t.rp.blk)
}

// puts currslot / recordpage to first record in file
func (t TableScan) beforeFirst() {
	t.moveToBlock(0)
}

// puts currRID at next record
func (t TableScan) next() bool {
	// temp var not strictly necessary
	nextslot := t.rp.nextAfter(t.currslot)
	for nextslot < 0 {
		t.moveToNextBlock() 
		nextslot = t.rp.nextAfter(nextslot)
	}
	t.currslot = nextslot

	return true
}

// this (currently) includes being able to move to nonexistent blocks (by creating them)
func (t TableScan) moveToRid(r RID) {
	t.currslot = r.slot
	if r.blknum != int(t.rp.blk.blknum) {
		t.moveToBlock(r.blknum)
	}
}

func (t TableScan) insert() {
	newslot := t.rp.insertAfter(t.currslot)
	for newslot < 0 {
		t.moveToNextBlock()
		// loop, etc, append block if all full, change currRID.blknum accordingly
		newslot = t.rp.insertAfter(t.currslot)
	}

	t.currslot = newslot
}

func (t TableScan) getInt(fldname string) int {
	return t.rp.getInt(t.currslot, fldname)
}
func (t TableScan) getString(fldname string) string {
	return t.rp.getString(t.currslot, fldname)
}
func (t TableScan) setInt(fldname string, val int) {
	t.rp.setInt(t.currslot, fldname, val)
}
func (t TableScan) setString(fldname string, val string) {
	t.rp.setString(t.currslot, fldname, val)
}

func (t TableScan) hasField(fldname string) bool {
	return t.layout.schema.hasField(fldname)
}

func (t TableScan) currentRID() RID {
	return RID{int(t.rp.blk.blknum), t.currslot}
}
func (t TableScan) delete() {
	t.rp.delete(t.currslot)
}

// aux

// having this isn't strictly needed, could just do moveToBlock() and calculate blknum up above
func (t TableScan) moveToNextBlock() {
	t.close()

	if t.atLastBlock() {
		t.moveToNewBlock()
	}

	newblknum := t.rp.blk.blknum + 1
	t.moveToBlock(newblknum)
}

func (t TableScan) moveToNewBlock() {
	t.close()
	newblk := t.tx.append(t.filename)

	newrp := makeRecordPage(t.tx, newblk, t.layout)
	newrp.format()

	t.rp = newrp
	t.currslot = -1
}

func (t TableScan) moveToBlock(newblknum int) {
	if newblknum == t.rp.blk.blknum {
		return
	}

	t.close()
	rp := makeRecordPage(t.tx, BlockId{t.filename, newblknum}, t.layout)
	
	t.rp = rp
	t.currslot = -1
}

func (t TableScan) atLastBlock() bool {
	return t.tx.size(t.filename) -1 == t.rp.blk.blknum 
}