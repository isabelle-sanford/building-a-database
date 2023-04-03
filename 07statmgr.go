package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

/**
Stats are:
B(T): number of blocks used by each table T
R(T): number of records in each table T
V(T,F): for each field F of table T, number of distinct F-values in T
*/

type StatMgr struct {
	tm *TableMgr
	//tx         *Transaction
	tableStats map[string]StatInfo
	numcalls   int
}

func (sm StatMgr) String() string {
	statTbl := fmt.Sprint("tblname	# records	# blocks")
	totrecs := 0
	totblocks := 0

	for tblname, si := range sm.tableStats {
		// TODO format tables better
		statTbl += fmt.Sprintf("%s	%d	%d\n", tblname, si.numrecs, si.numblocks)
		totrecs += si.numrecs
		totblocks += si.numblocks
	}

	ret := fmt.Sprint("DATABASE STATS---\n")
	ret += fmt.Sprintf(
		"%d tables containing %d records in %d blocks [more here?]\n (stats last refreshed %d calls ago)\n", len(sm.tableStats), totrecs, totblocks, sm.numcalls)

	return ret + statTbl
}

type StatInfo struct {
	tblname   string
	numblocks int
	numrecs   int
	distincts map[string]int
}

func (si StatInfo) String() string {
	return fmt.Sprintf("Stats (%s): %d record(s) in %d block(s)", si.tblname, si.numrecs, si.numblocks)
}

func makeStatMgr(tm *TableMgr, tx *Transaction) StatMgr {

	sm := StatMgr{tm, make(map[string]StatInfo, 0), 0}
	sm.refreshStatistics(tx)

	return sm
}

// might need layout arg? not sure why tho
// don't even really need tx
// ! also maybe doesn't need to be pointered?
func (sm *StatMgr) getStatInfo(tblname string, layout Layout, tx *Transaction) *StatInfo {
	sm.numcalls++
	if sm.numcalls > 10 { // ! temp for testing
		sm.refreshStatistics(tx)
	}

	si, ok := sm.tableStats[tblname]

	if !ok {
		si = calcTableStats(tblname, layout, tx)
		sm.tableStats[tblname] = si
	}

	return &si
}

func (sm *StatMgr) refreshStatistics(tx *Transaction) {
	sm.numcalls = 0
	sm.tableStats = make(map[string]StatInfo)
	tcat := makeTableScan(tx, "tblcat", sm.tm.tblcat) // probably shouldn't hard-code like that

	for tcat.next() {
		tblname, _ := tcat.getString("tblname")
		layout := sm.tm.getLayout(tblname, tx)
		si := calcTableStats(tblname, layout, tx)
		sm.tableStats[tblname] = si
	}
	tcat.close() // I forget what this does

}

// need an if-not-exists (though always should?)
func calcTableStats(tblname string, layout Layout, tx *Transaction) StatInfo {

	tbl := makeTableScan(tx, tblname, layout)

	numrecs := 0
	numblocks := 0

	for tbl.next() {
		numrecs++
		numblocks = tbl.getRid().blknum + 1 // plus 1 for zero-indexing
	}

	distincts := make(map[string]int) // todo

	tbl.close()

	return StatInfo{tblname, numblocks, numrecs, distincts}
}

func (si *StatInfo) getDistinct(fldname string) int {
	return 1 + si.numrecs/3 // nope!
}

func StatTest() {
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

	logfile, _ := os.Create("dblogs.txt")
	logger := log.New(logfile, "logger: ", log.Lshortfile)
	tm.showTblCatalog(logger)

	fmt.Println(sm)

	si := sm.getStatInfo("Table1", layout, tx)

	fmt.Println(sm)

	fmt.Printf("%d blocks, %d recs, %d distinct in A\n", si.numblocks, si.numrecs, si.getDistinct("A"))

	si2 := sm.getStatInfo("Table1", layout, tx)

	fmt.Println(si2)

}
