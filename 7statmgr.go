package main

import (
	"fmt"
	"math/rand"
)

/**
Stats are:
B(T): number of blocks used by each table T
R(T): number of records in each table T
V(T,F): for each field F of table T, number of distinct F-values in T
*/

type StatMgr struct {
	tm         *TableMgr
	tx         *Transaction
	tableStats map[string]StatInfo
	numcalls   int
}

type StatInfo struct {
	sm        *StatMgr
	tblname   string
	numblocks int
	numrecs   int
	distincts map[string]int
}

func makeStatMgr(tm *TableMgr, tx *Transaction) StatMgr {
	tblcat := makeTableScan(tx, "tblcat", tm.tblcat) // probably shouldn't hard-code like that

	tableStats := make(map[string]StatInfo, 0)

	sm := StatMgr{tm, tx, tableStats, 0}

	for tblcat.next() {
		tblname := tblcat.getString("tblname")

		tableStats[tblname] = *sm.makeStatInfo(tblname)
	}

	return sm
}

func (sm *StatMgr) makeStatInfo(tblname string) *StatInfo {

	// could put fields in here?
	si := StatInfo{sm, tblname, -1, -1, make(map[string]int)} // NO (offsets is not # distinct values ???)

	si.refreshTableStatistics()

	sm.tableStats[tblname] = si

	return &si

}

// need an if-not-exists
func (si *StatInfo) refreshTableStatistics() {
	tbl := makeTableScan(si.sm.tx, si.tblname, si.sm.tm.getLayout(si.tblname, si.sm.tx))

	numrecs := 0

	for tbl.next() {
		numrecs++
	}

	si.numblocks = tbl.currentRID().blknum

	si.numrecs = numrecs

	// ! add distincts
}

func (si *StatInfo) getDistinct(fldname string) int {
	return 1 + si.numrecs/3 // nope!
}

func (sm *StatMgr) getStatInfo(tblname string, layout Layout, tx Transaction) {
	sm.numcalls++
	if sm.numcalls > 100 {
		sm.re
	}
}

func main() {
	db := makeDB()
	tx := db.makeTx()
	tm := makeTableMgr(tx, true)
	sm := makeStatMgr(&tm, tx)

	var sch Schema = makeSchema()

	sch.addIntField("A")
	sch.addStringField("B", 9)
	layout := makeLayoutFromSchema(sch)
	for fldname := range layout.schema.fields {
		offset := layout.offsets[fldname]
		fmt.Printf("%s had offset %d\n", fldname, offset)
	}
	fmt.Printf("Total slot size is %d\n", layout.slotsize)

	ts := makeTableScan(tx, "Table1", layout)

	//fmt.Println(ts.rp.tx.bufflist.buffers[BlockId{"NewT.tbl", 0}].pg)

	fmt.Println("Filling the page with random records.")
	ts.beforeFirst()
	for i := 0; i < 20; i++ {
		ts.insert() // !
		n := rand.Intn(50)
		ts.setInt("A", n)
		ts.setString("B", fmt.Sprint("rec", n))
		//fmt.Printf("Inserting into slot %v: {%d, %s%d}\n", ts.currentRID(), n, "rec", n)
	}

	tm.createTable("Table1", sch, tx)

	tm.showTblCatalog()

	fmt.Println(sm)

	sm.makeStatInfo("Table1")

	si := sm.tableStats["Table1"]

	fmt.Printf("%d %d %d\n", si.numblocks, si.numrecs, si.getDistinct("A"))

}
