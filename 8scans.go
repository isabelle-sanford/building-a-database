package main

type Predicate struct {
	// TODO
}

// used by: TableScan, SelectScan, ProjectScan, ProductScan
type Scan interface {
	beforeFirst()
	next() bool
	getInt(fldname string) int
	getString(fldname string) int
	//getVal(fldname string) // need to figure out equivalent of "Constant" in go
	hasField(fldname string) bool
	close()
}

// used by TableScan and SelectScan
type UpdateScan interface {
	Scan
	setInt(fldname string, val int)
	setString(fldname string, val string)
	//getVal(fldname string, val CONST) // (see in Scan)
	insert()
	delete()

	getRID() RID
	moveToRID(rid RID)
}

type SelectScan struct {
	scn  Scan // can this be not pointered without problems? interface doesn't work without it
	pred Predicate
}

// func makeSelectScan(scnIn *Scan, pred Predicate) *SelectScan {

// }

func (ss *SelectScan) beforeFirst() {
	ss.scn.beforeFirst()
}
