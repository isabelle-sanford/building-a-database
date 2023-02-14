package main

// ! okay but just make these anonymous tho?
type MetadataMgr struct {
	// ? should these still be poitners
	*TableMgr
	*StatMgr
}

func makeMetadataMgr(isnew bool, tx *Transaction) *MetadataMgr {
	tm := makeTableMgr(tx, isnew)
	sm := makeStatMgr(&tm, tx)

	mdmgr := MetadataMgr{&tm, &sm}

	return &mdmgr
}

// func (mdmgr *MetadataMgr) createTable(tblname String, sch Schema, tx *Transaction) {
// 	mdmgr.tblmgr.createTable(tblname, sch, tx)
// }
