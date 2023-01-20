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
	sm        *StatMgr
	tblname   string
	numblocks int
	numrecs   int
	distincts map[string]int
}

func (sm *StatMgr) makeStatInfo(tblname string) *StatInfo {

	// could put fields in here?
	si := StatInfo{sm, tblname, -1, -1, make(map[string]int)} // NO (offsets is not # distinct values ???)

	si.refreshTableStatistics()

	return &si

}

func (si StatInfo) refreshTableStatistics() {
	tbl := makeTableScan(si.sm.tx, si.tblname, si.sm.tm.getLayout(si.tblname, si.sm.tx))

	numrecs := 0

	for tbl.next() {
		numrecs++
	}

	si.numblocks = tbl.currentRID().blknum

	si.numrecs = numrecs
}

func main() {

}
