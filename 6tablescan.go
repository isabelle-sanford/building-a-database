package main

import (
	"fmt"
	"math/rand"
)

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

func makeTableScan(tx *Transaction, tblname string, layout Layout) TableScan {
	// MAKE file (bc files are homogeneous)
	filename := tblname + ".tbl"

	// might be immediately overwritten by the moveToNewBlock()
	rp := makeRecordPage(tx, BlockId{filename, 0}, layout)

	t := TableScan{tx, tblname, layout, &rp, filename, 0}

	// fmt.Println(t)
	if tx.fm.openFiles[filename] > 0 {
		t.moveToBlock(0)
	} else {
		t.moveToNewBlock()
	}

	return t
}

func (t *TableScan) close() {
	t.tx.unpin(t.rp.blk)
}

// puts currslot / recordpage to first record in file
func (t *TableScan) beforeFirst() {
	t.moveToBlock(0)
}

// puts currRID at next record
func (t *TableScan) next() bool {
	// temp var not strictly necessary
	nextslot := t.rp.nextAfter(t.currslot)
	for nextslot < 0 {
		if t.atLastBlock() {
			return false
		}
		t.moveToNextBlock()
		nextslot = t.rp.nextAfter(nextslot)
	}
	t.currslot = nextslot

	return true
}

// this (currently) includes being able to move to nonexistent blocks/records (by creating them)
func (t *TableScan) moveToRid(r RID) {
	t.currslot = r.slot
	if r.blknum != int(t.rp.blk.blknum) {
		t.moveToBlock(r.blknum)
	}
}

func (t *TableScan) insert() {
	newslot := t.rp.insertAfter(t.currslot)
	for newslot < 0 {
		t.moveToNextBlock()
		// loop, etc, append block if all full, change currRID.blknum accordingly
		newslot = t.rp.insertAfter(t.currslot)
	}

	t.currslot = newslot
}

func (t *TableScan) getInt(fldname string) int {
	return t.rp.getInt(t.currslot, fldname)
}
func (t *TableScan) getString(fldname string) string {
	return t.rp.getString(t.currslot, fldname)
}
func (t *TableScan) setInt(fldname string, val int) {
	t.rp.setInt(t.currslot, fldname, val)
}
func (t *TableScan) setString(fldname string, val string) {
	t.rp.setString(t.currslot, fldname, val)
}

func (t *TableScan) hasField(fldname string) bool {
	return t.layout.schema.hasField(fldname)
}

func (t *TableScan) currentRID() RID {
	return RID{int(t.rp.blk.blknum), t.currslot}
}
func (t *TableScan) delete() {
	t.rp.delete(t.currslot)
}

// aux

// having this isn't strictly needed(?), could just do moveToBlock() and calculate blknum up above
func (t *TableScan) moveToNextBlock() {
	t.close()

	if t.atLastBlock() {
		t.moveToNewBlock()
	} else {
		newblknum := t.rp.blk.blknum + 1
		t.moveToBlock(newblknum)
	}

}

func (t *TableScan) moveToNewBlock() {
	t.close()
	newblk := t.tx.append(t.filename)

	newrp := makeRecordPage(t.tx, newblk, t.layout)
	newrp.format()

	t.rp = &newrp
	t.currslot = -1

}

func (t *TableScan) moveToBlock(newblknum int) {
	// should newblknum be set into t somewhere here?

	// if newblknum == t.rp.blk.blknum {
	// 	return
	// }

	t.close()
	rp := makeRecordPage(t.tx, BlockId{t.filename, newblknum}, t.layout)

	t.rp = &rp
	t.currslot = -1
}

func (t *TableScan) atLastBlock() bool {
	return t.tx.fm.openFiles[t.filename]-1 == t.rp.blk.blknum
}

// tableScanTest
func main() {
	vfm := makeFileMgr("mydb", 400)
	fm := &vfm
	vlm := makeLogMgr(fm, "log")
	lm := &vlm
	vbm := makeBufferManager(fm, lm, 8)
	bm := &vbm

	tx := makeTransaction(fm, lm, bm)
	var sch Schema = makeSchema()

	sch.addIntField("A")
	sch.addStringField("B", 9)
	layout := makeLayoutFromSchema(sch)
	for fldname := range layout.schema.fields {
		offset := layout.offsets[fldname]
		fmt.Printf("%s had offset %d\n", fldname, offset)
	}
	fmt.Printf("Total slot size is %d\n", layout.slotsize)

	ts := makeTableScan(tx, "NewT", layout) // !

	fmt.Println("Filling the page with random records.")
	ts.beforeFirst()
	for i := 0; i < 25; i++ {
		ts.insert()
		n := rand.Intn(50)
		ts.setInt("A", n)
		ts.setString("B", fmt.Sprint("rec", n))
		fmt.Printf("Inserting into slot %v: {%d, %s%d}\n", ts.currentRID(), n, "rec", n)
	}

	fmt.Println("Deleting these records with A values < 25.")
	count := 0
	ts.beforeFirst()
	for ts.next() {
		a := ts.getInt("A")
		b := ts.getString("B")
		if a < 25 {
			count++
			fmt.Printf("Slot %v: {%d, %s}\n", ts.currentRID(), a, b)
			ts.delete()
		}
	}

	fmt.Printf("%d values under 25 were deleted.\n", count)
	fmt.Println("Here are the remining records.")
	ts.beforeFirst()
	for ts.next() {
		a := ts.getInt("A")
		b := ts.getString("B")
		fmt.Printf("Slot %v: {%d, %s}\n", ts.currentRID(), a, b)
	}

	ts.close()
	tx.commit()
}
