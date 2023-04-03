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

	t := TableScan{tx: tx, tblname: tblname, layout: layout, filename: filename}

	if tx.fm.openFiles[filename] > 0 {
		//fmt.Println("file already has stuff in it!")
		//fmt.Println("has stuff; openFiles looks like: ", tx.fm.openFiles)
		t.moveToBlock(0)
	} else {
		t.moveToNewBlock()
	}

	return t
}

func (t *TableScan) close() {
	if t.rp != nil {
		t.tx.unpin(t.rp.blk)
	}

}

// puts currslot / recordpage to first record in file
func (t TableScan) beforeFirst() {
	t.moveToBlock(0)
}

// puts currRID at next record
func (t *TableScan) next() bool {
	t.currslot = t.rp.nextAfter(t.currslot)
	for t.currslot < 0 {
		if t.atLastBlock() {
			return false
		}
		t.moveToBlock(t.rp.blk.blknum + 1)
		t.currslot = t.rp.nextAfter(t.currslot)
	}

	return true
}

// this (currently) includes being able to move to nonexistent blocks/records (by creating them)
func (t *TableScan) moveToRid(r RID) {
	t.currslot = r.slot
	if r.blknum != int(t.rp.blk.blknum) {
		t.moveToBlock(r.blknum)
	}
}

func (t *TableScan) getVal(fldname string) (Constant, bool) {

	typ := t.layout.schema.fields[fldname]
	if typ.fldtype == INTEGER {
		val := t.rp.getInt(t.currslot, fldname)
		return makeConstInt(val), true
	} else if typ.fldtype == VARCHAR {
		val := t.rp.getString(t.currslot, fldname)
		return makeConstString(val), true
	} else {
		return Constant{}, false // i guess
	}
}
func (t *TableScan) setVal(fldname string, val Constant) {
	typ := t.layout.schema.fields[fldname]
	if typ.fldtype == INTEGER {
		t.rp.setInt(t.currslot, fldname, val.ival)

	} else if typ.fldtype == VARCHAR {
		t.rp.setString(t.currslot, fldname, val.sval)
	}
}

func (t *TableScan) getInt(fldname string) (int, bool) {
	return t.rp.getInt(t.currslot, fldname), true // i guess
}
func (t *TableScan) getString(fldname string) (string, bool) {
	return t.rp.getString(t.currslot, fldname), true
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

func (t *TableScan) getRid() RID {
	return RID{int(t.rp.blk.blknum), t.currslot}
}
func (t *TableScan) delete() {
	t.rp.delete(t.currslot)
}
func (t *TableScan) insert() {
	//fmt.Println("Looking at currblock ", t.rp.blk)
	newslot := t.rp.insertAfter(t.currslot)
	//fmt.Println("Next slot from insertAfter is ", newslot)
	for newslot < 0 {
		if t.atLastBlock() {
			t.moveToNewBlock()
		} else {
			t.moveToBlock(t.rp.blk.blknum + 1)
		}

		// loop, etc, append block if all full, change currRID.blknum accordingly
		newslot = t.rp.insertAfter(t.currslot)
	}
	//fmt.Println("Setting currslot to ", newslot)

	t.currslot = newslot
}

// aux

func (t *TableScan) moveToNewBlock() {
	t.close()
	newblk := t.tx.append(t.filename)

	newrp := makeRecordPage(t.tx, newblk, t.layout)
	newrp.format()

	t.rp = &newrp
	t.currslot = -1

}

func (t *TableScan) moveToBlock(newblknum int) {

	t.close()
	rp := makeRecordPage(t.tx, BlockId{t.filename, newblknum}, t.layout)

	t.rp = &rp
	t.currslot = -1
}

func (t *TableScan) atLastBlock() bool {
	return t.tx.fm.openFiles[t.filename]-1 == t.rp.blk.blknum
}

func (ts TableScan) String() string {
	ts.beforeFirst()

	ret := fmt.Sprintf("Table %s: \n%v\n", ts.tblname, ts.layout.schema)

	for ts.next() {
		ret += ts.printRecord() + "\n"
	}

	return ret + "--\n"
}

func (ts *TableScan) printRecord() string {
	var s string
	var val string
	for _, fldname := range ts.layout.schema.fieldlist {
		fld := ts.rp.layout.schema.fields[fldname]
		if fld.fldtype == VARCHAR {
			val, _ = ts.getString(fldname)
		} else {
			temp, _ := ts.getInt(fldname)
			val = fmt.Sprint(temp)
		}

		s += val + " "
	}
	return s
}

// tableScanTest
func tableScanTest() {
	vfm := makeFileMgr("mydb", 400)
	fm := &vfm
	vlm := makeLogMgr(fm, "log")
	lm := &vlm
	vbm := makeBufferManager(fm, lm, 9)
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

	ts := makeTableScan(tx, "NewTable1", layout)

	//fmt.Println(ts.rp.tx.bufflist.buffers[BlockId{"NewT.tbl", 0}].pg)

	fmt.Println("Filling the page with random records.")
	ts.beforeFirst()
	for i := 0; i < 20; i++ {
		ts.insert() // !
		n := rand.Intn(50)
		ts.setInt("A", n)
		ts.setString("B", fmt.Sprint("rec", n))
		fmt.Printf("Inserting into slot %v: {%d, %s%d}\n", ts.getRid(), n, "rec", n)
	}

	//fmt.Println(ts.rp.tx.bufflist.buffers[BlockId{"NewT.tbl", 0}].pg)

	fmt.Println("Deleting these records with A values < 25.")
	count := 0
	ts.beforeFirst()
	for ts.next() {
		a, _ := ts.getInt("A")
		b, _ := ts.getString("B")
		if a < 25 {
			count++
			fmt.Printf("Slot %v: {%d, %s}\n", ts.getRid(), a, b)
			ts.delete()
		}
	}

	fmt.Printf("%d values under 25 were deleted.\n", count)
	fmt.Println("Here are the remining records.")
	ts.beforeFirst()
	for ts.next() {
		a, _ := ts.getInt("A")
		b, _ := ts.getString("B")
		fmt.Printf("Slot %v: {%d, %s}\n", ts.getRid(), a, b)
	}

	ts.close()
	tx.commit()
}
