package main

import (
	"fmt"
	"math/rand"
)

type RecordPage struct {
	tx     *Transaction
	blk    BlockId
	layout Layout
}

const EMPTY int = 1
const USED int = 2

// ? consider returning *RecordPage
func makeRecordPage(tx *Transaction, blk BlockId, layout Layout) RecordPage {
	rp := RecordPage{tx, blk, layout}
	tx.pin(blk)
	return rp
}

func (rp *RecordPage) getInt(slot int, fldname string) int {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	return rp.tx.getInt(rp.blk, fldpos)
}

func (rp *RecordPage) getString(slot int, fldname string) string {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	return rp.tx.getString(rp.blk, fldpos)
}

func (rp *RecordPage) setInt(slot int, fldname string, val int) {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	rp.tx.setInt(rp.blk, fldpos, val, true)
}

func (rp *RecordPage) setString(slot int, fldname string, val string) {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	rp.tx.setString(rp.blk, fldpos, val, true)
}

func (rp *RecordPage) delete(slot int) {
	rp.setFlag(slot, EMPTY)
}

func (rp *RecordPage) format() {
	slot := 0
	for rp.isValidSlot(slot) {
		rp.tx.setInt(rp.blk, rp.offset(slot), EMPTY, false)
		sch := rp.layout.schema
		for f := range sch.fields {
			fldpos := rp.offset(slot) + rp.layout.offsets[f]
			if sch.fldtype(f) == INTEGER {
				rp.tx.setInt(rp.blk, fldpos, 0, false)
			} else {
				rp.tx.setString(rp.blk, fldpos, "", false)
			}
		}
		slot++
	}
	//fmt.Println("RP just post-formatting: ", rp.tx.bufflist.buffers[rp.blk].pg)
}

// finds next used slot after @param slot
func (rp *RecordPage) nextAfter(slot int) int {
	return rp.searchAfter(slot, USED)
}

func (rp *RecordPage) insertAfter(slot int) int {
	//fmt.Println("RP at start of insertAfter: ", rp.tx.bufflist.buffers[rp.blk].pg)
	newslot := rp.searchAfter(slot, EMPTY)
	//fmt.Println("insertAfter gets from searchAfter: \n", newslot)
	if newslot >= 0 {
		rp.setFlag(newslot, USED)
	}
	return newslot
}

// "private" aux methods

func (rp *RecordPage) setFlag(slot int, flag int) {
	rp.tx.setInt(rp.blk, rp.offset(slot), flag, true)
}

func (rp *RecordPage) searchAfter(slot int, flag int) int {
	sslot := slot + 1
	//fmt.Println("searchAfter checks whether this slot is valid: ", sslot)
	for rp.isValidSlot(sslot) {
		fmt.Printf("Slot is valid! Checking flag at block %v offset %d: %d \n", rp.blk, rp.offset(sslot), rp.tx.getInt(rp.blk, rp.offset(sslot)))

		if rp.tx.getInt(rp.blk, rp.offset(sslot)) == flag {
			//fmt.Println("Slot is unused: ", sslot)
			return sslot
		}
		sslot++
	}
	return -1
}

func (rp *RecordPage) isValidSlot(slot int) bool {
	return rp.offset(slot+1) <= rp.tx.fm.blocksize
}

func (rp *RecordPage) offset(slot int) int {
	return slot * rp.layout.slotsize
}

func recordTest() {
	vfm := makeFileMgr("mydb", 400)
	fm := &vfm
	vlm := makeLogMgr(fm, "log")
	lm := &vlm
	vbm := makeBufferManager(fm, lm, 8)
	bm := &vbm

	//p := makePage(fm.blocksize)

	tx1 := makeTransaction(fm, lm, bm)

	sch := makeSchema()
	sch.addIntField("A")
	sch.addStringField("B", 9)

	layout := makeLayoutFromSchema(sch)

	for fldname := range layout.schema.fields {
		offset := layout.offsets[fldname]
		fmt.Printf("%s had offset %d\n", fldname, offset)
	}
	fmt.Printf("Total slot size is %d\n", layout.slotsize)

	blk := tx1.append("testfile")
	tx1.pin(blk)
	rp := makeRecordPage(tx1, blk, layout)
	rp.format()

	fmt.Println("Page now looks like: ", rp.tx.bufflist.buffers[blk].pg)

	fmt.Println("Filling the page with random records.")
	slot := rp.insertAfter(-1)

	for slot >= 0 {
		n := rand.Intn(50)
		rp.setInt(slot, "A", n)
		rp.setString(slot, "B", fmt.Sprint("rec", n))
		fmt.Printf("Inserting into slot %d: {%d, %s%d}\n", slot, n, "rec", n)
		slot = rp.insertAfter(slot)
	}

	fmt.Println("Page now looks like: \n", rp.tx.bufflist.buffers[blk].pg)

	fmt.Println("Deleting these records with A values < 25.")
	count := 0
	slot = rp.nextAfter(-1)
	for slot >= 0 {
		a := rp.getInt(slot, "A")
		b := rp.getString(slot, "B")
		if a < 25 {
			count++
			fmt.Printf("Slot %d: {%d, %s}\n", slot, a, b)
			rp.delete(slot)
		}
		slot = rp.nextAfter(slot)
	}
	fmt.Printf("%d values under 25 were deleted.\n", count)
	fmt.Println("Here are the remining records.")
	slot = rp.nextAfter(-1)
	for slot >= 0 {
		a := rp.getInt(slot, "A")
		b := rp.getString(slot, "B")
		fmt.Printf("slot %d: {%d, %s}\n", slot, a, b)
		slot = rp.nextAfter(slot)
	}

	tx1.unpin(blk)
	tx1.commit()
}
