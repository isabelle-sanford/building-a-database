package main

/**
Stats are:
B(T): number of blocks used by each table T
R(T): number of records in each table T
V(T,F): for each field F of table T, number of distinct F-values in T
*/

type StatMgr struct {
	tm         *TableMgr
	tx         *Transaction
	tableStats []StatInfo
}

type StatInfo struct {
	//tx *Transaction
	tblname   string
	numblocks int
	numrecs   int
	distincts map[string]int
}

func (sm *StatMgr) makeStatInfo(tblname string) *StatInfo {
	tbl := makeTableScan(sm.tx, tblname, sm.tm.getLayout(tblname, sm.tx))

	numrecs := 0

	for tbl.next() {
		numrecs++
	}

	numblocks := tbl.currentRID().blknum

	si := StatInfo{tblname, numblocks, numrecs, tbl.layout.offsets} // NO (offsets is not # distinct values ???)

	return &si

}

func (si StatInfo) refreshTableStatistics() {

}
