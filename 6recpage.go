package main

type RecordPage struct {
	tx     Transaction
	blk    BlockId
	layout Layout
}

const EMPTY int = 0
const USED int = 1

func makeRecordPage(tx Transaction, blk BlockId, layout Layout) RecordPage {
	rp := RecordPage{tx, blk, layout}
	tx.pin(blk)
	return rp
}

func (rp RecordPage) getInt(slot int, fldname string) int {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	return rp.tx.getInt(rp.blk, fldpos)
}

func (rp RecordPage) getString(slot int, fldname string) string {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	return rp.tx.getString(rp.blk, fldpos)
}

func (rp RecordPage) setInt(slot int, fldname string, val int) {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	rp.tx.setInt(rp.blk, fldpos, val, true)
}

func (rp RecordPage) setString(slot int, fldname string, val string) {
	fldpos := rp.offset(slot) + rp.layout.offsets[fldname]
	rp.tx.setString(rp.blk, fldpos, val, true)
}

func (rp RecordPage) delete(slot int) {
	rp.setFlag(slot, EMPTY)
}

func (rp RecordPage) format() {
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
}

// finds next used slot after @param slot
func (rp RecordPage) nextAfter(slot int) int {
	return rp.searchAfter(slot, USED)
}

func (rp RecordPage) insertAfter(slot int) int {
	newslot := rp.searchAfter(slot, EMPTY)
	if newslot >= 0 {
		rp.setFlag(newslot, USED)
	}
	return newslot
}

// "private" aux methods




func (rp RecordPage) setFlag(slot int, flag int) {
	rp.tx.setInt(rp.blk, rp.offset(slot), flag, true)
}

func (rp RecordPage) searchAfter(slot int, flag int) int {
	slot++
	for rp.isValidSlot(slot) {
		if rp.tx.getInt(rp.blk, rp.offset(slot)) == flag {
			return slot
		}
		slot++
	}
	return -1
}

func (rp RecordPage) isValidSlot(slot int) bool {
	return rp.offset(slot + 1) <= rp.tx.blockSize()
}

func (rp RecordPage) offset(slot int) int {
	return slot * rp.layout.slotsize
}
