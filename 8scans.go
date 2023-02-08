package main

// used by: TableScan, SelectScan, ProjectScan, ProductScan
type Scan interface {
	beforeFirst()
	next() bool
	getInt(fldname string) int
	getString(fldname string) string
	getVal(fldname string) Constant // need to figure out equivalent of "Constant" in go
	hasField(fldname string) bool
	close()
}

// used by TableScan and SelectScan
type UpdateScan interface {
	Scan
	setInt(fldname string, val int)
	setString(fldname string, val string)
	setVal(fldname string, val Constant) // (see in Scan)
	insert()
	delete()

	getRID() RID
	moveToRID(rid RID)
}

// SELECT SCAN

type SelectScan struct {
	// is this ok? just because SelectScan implements updatescan doesn't mean scn does?
	scn  UpdateScan // can this be not pointered without problems? interface doesn't work without it
	pred Predicate
}

func (ss *SelectScan) beforeFirst() {
	ss.scn.beforeFirst()
}

func (ss *SelectScan) next() bool {
	for ss.scn.next() {
		if ss.pred.isSatisfied(ss) {
			return true
		}
	}
	return false
}

func (ss *SelectScan) getInt(fldname string) int {
	return ss.scn.getInt(fldname)
}

func (ss *SelectScan) getString(fldname string) string {
	return ss.scn.getString(fldname)
}

func (ss *SelectScan) getVal(fldname string) Constant {
	return ss.scn.getVal(fldname)
}

func (ss *SelectScan) hasField(fldname string) bool {
	return ss.scn.hasField(fldname)
}

func (ss *SelectScan) close() {
	ss.scn.close()
}

// for UpdateScan
func (ss *SelectScan) setInt(fldname string, val int) {
	ss.scn.setInt(fldname, val)
}

func (ss *SelectScan) setString(fldname string, val string) {
	ss.scn.setString(fldname, val)
}

func (ss *SelectScan) setVal(fldname string, val Constant) {
	ss.scn.setVal(fldname, val)
}

func (ss *SelectScan) delete() {
	ss.scn.delete()
}

func (ss *SelectScan) insert() {
	ss.scn.insert()
}

func (ss *SelectScan) getRID() {
	ss.scn.getRID()
}

func (ss *SelectScan) moveToRID(rid RID) {
	ss.scn.moveToRID(rid)
}

// PROJECT SCAN

type ProjectScan struct {
	scn       Scan           // can this be not pointered without problems? interface doesn't work without it
	fieldlist map[string]int // slice vs map ?
}

func (ps *ProjectScan) beforeFirst() {
	ps.scn.beforeFirst()
}

func (ps *ProjectScan) next() bool {
	return ps.scn.next()
}

func (ps *ProjectScan) getInt(fldname string) int {
	if ps.hasField(fldname) {
		return ps.scn.getInt(fldname)
	}
	// ! throw runtime exception field not found
}

func (ps *ProjectScan) getString(fldname string) string {
	if ps.hasField(fldname) {
		return ps.scn.getString(fldname)
	}
	// ! throw runtime exception field not found
}

func (ps *ProjectScan) getVal(fldname string) Constant {
	if ps.hasField(fldname) {
		return ps.scn.getVal(fldname)
	}
	// throw runtime exception

}

func (ps *ProjectScan) hasField(fldname string) bool {
	_, ok := ps.fieldlist[fldname]
	return ok
}

func (ps *ProjectScan) close() {
	ps.scn.close()
}

type ProductScan struct {
	s1 Scan
	s2 Scan
}

func (ps *ProductScan) beforeFirst() {
	ps.s1.beforeFirst()
	ps.s1.next()
	ps.s2.beforeFirst()
}

func (ps *ProductScan) next() bool {
	if ps.s2.next() {
		return true
	} else {
		ps.s2.beforeFirst()
		return ps.s2.next() && ps.s1.next()
	}
}

func (ps *ProductScan) getInt(fldname string) int {
	if ps.s1.hasField(fldname) {
		return ps.s1.getInt(fldname)
	} else {
		return ps.s2.getInt(fldname)
	}
	// ! throw runtime exception field not found
}

func (ps *ProductScan) getString(fldname string) string {
	if ps.s1.hasField(fldname) {
		return ps.s1.getString(fldname)
	} else {
		return ps.s2.getString(fldname)
	}
}

func (ps *ProductScan) getVal(fldname string) Constant {
	if ps.s1.hasField(fldname) {
		return ps.s1.getVal(fldname)
	} else {
		return ps.s2.getVal(fldname)
	}
}

func (ps *ProductScan) hasField(fldname string) bool {
	return ps.s1.hasField(fldname) || ps.s2.hasField(fldname)
}

func (ps *ProductScan) close() {
	ps.s1.close()
	ps.s2.close()
}