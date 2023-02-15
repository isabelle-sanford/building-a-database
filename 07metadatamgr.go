package main

// ! okay but just make these anonymous tho?
type MetadataMgr struct {
	// ? should these still be poitners
	tm *TableMgr
	sm *StatMgr
}

func makeMetadataMgr(isnew bool, tx *Transaction) *MetadataMgr {
	tm := makeTableMgr(tx, isnew)
	sm := makeStatMgr(&tm, tx)

	mdm := MetadataMgr{&tm, &sm}

	return &mdm
}

func (mdm *MetadataMgr) createTable(tblname string, sch Schema, tx *Transaction) {
	mdm.tm.createTable(tblname, sch, tx)
}

func (mdm *MetadataMgr) getLayout(tblname string, tx *Transaction) Layout {
	return mdm.tm.getLayout(tblname, tx)
}

func (mdm *MetadataMgr) getStatInfo(tblname string, layout Layout, tx *Transaction) StatInfo {
	return *mdm.sm.getStatInfo(tblname, layout, tx)
}

func (mdm *MetadataMgr) createIndex(idxname string, tblname string, fldname string, tx *Transaction) {
	return
}

func (mdm *MetadataMgr) getIndexInfo(tblname string, tx *Transaction) {
	return // should return map[string]IndexInfo so that's fun
}

// not sure abt viewdef type
func (mdm *MetadataMgr) createView(viewname string, viewdef string, tx *Transaction) {
	return
}
func (mdm *MetadataMgr) getViewDef(viewname string, tx *Transaction) { // should return string
	return
}
