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

func (mdmgr *MetadataMgr) createIndex(idxname string, tblname string, fldname string, tx *Transaction)

// not sure abt viewdef type
func (mdmgr *MetadataMgr) createView(viewname string, viewdef string, tx *Transaction)
