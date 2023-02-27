package main

import "fmt"

// used by: TableScan, SelectScan, ProjectScan, ProductScan
type Scan interface {
	beforeFirst()
	next() bool
	getInt(fldname string) (int, bool)
	getString(fldname string) (string, bool)
	getVal(fldname string) (Constant, bool) // need to figure out equivalent of "Constant" in go
	hasField(fldname string) bool
	close()
	String() string
}

// used by TableScan and SelectScan
type UpdateScan interface {
	Scan
	setInt(fldname string, val int)
	setString(fldname string, val string)
	setVal(fldname string, val Constant) // (see in Scan)
	insert()
	delete()

	getRid() RID
	moveToRid(rid RID)
}

// SELECT SCAN

type SelectScan struct {
	// is this ok? just because SelectScan implements updatescan doesn't mean scn does?
	scn  Scan // can this be not pointered without problems? interface doesn't work without it
	pred Predicate
}

func (ss SelectScan) String() string {
	return fmt.Sprint("select (WHERE)", ss.pred, " => \n", ss.scn)
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

func (ss *SelectScan) getInt(fldname string) (int, bool) {
	return ss.scn.getInt(fldname)
}

func (ss *SelectScan) getString(fldname string) (string, bool) {
	return ss.scn.getString(fldname)
}

func (ss *SelectScan) getVal(fldname string) (Constant, bool) {
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
	us := ss.scn.(UpdateScan)
	us.setInt(fldname, val)
}
func (ss *SelectScan) setString(fldname string, val string) {
	us := ss.scn.(UpdateScan)
	us.setString(fldname, val)
}
func (ss *SelectScan) setVal(fldname string, val Constant) {
	us := ss.scn.(UpdateScan)
	us.setVal(fldname, val)
}
func (ss *SelectScan) delete() {
	us := ss.scn.(UpdateScan)
	us.delete()
}
func (ss *SelectScan) insert() {
	us := ss.scn.(UpdateScan)
	us.insert()
}
func (ss *SelectScan) getRid() {
	us := ss.scn.(UpdateScan)
	us.getRid()
}
func (ss *SelectScan) moveToRid(rid RID) {
	us := ss.scn.(UpdateScan)
	us.moveToRid(rid)
}

// PROJECT SCAN---------------------------------

type ProjectScan struct {
	scn       Scan     // can this be not pointered without problems? interface doesn't work without it
	fieldlist []string // slice vs map ?
}

func (ps ProjectScan) String() string {
	return fmt.Sprint("Projecting/filtering to (SELECT ...) columns: ", ps.fieldlist, " => \n", ps.scn)
}

func (ps *ProjectScan) beforeFirst() {
	ps.scn.beforeFirst()
}
func (ps *ProjectScan) next() bool {
	return ps.scn.next()
}
func (ps *ProjectScan) getInt(fldname string) (int, bool) {
	var b bool
	if ps.hasField(fldname) {
		b = true
		ret, _ := ps.scn.getInt(fldname)
		return ret, b
	}
	return -1, b
	// ! throw runtime exception field not found
}
func (ps *ProjectScan) getString(fldname string) (string, bool) {
	var b bool
	if ps.hasField(fldname) {
		b = true
		ret, _ := ps.scn.getString(fldname)
		return ret, b
	}
	return "", b
	// ! throw runtime exception field not found
}
func (ps *ProjectScan) getVal(fldname string) (Constant, bool) {
	var b bool
	if ps.hasField(fldname) {
		b = true
		ret, _ := ps.scn.getVal(fldname)
		return ret, b
	}

	return Constant{}, b
	// throw runtime exception
}
func (ps *ProjectScan) hasField(fldname string) bool {
	for _, k := range ps.fieldlist {
		if k == fldname {
			return true
		}
	}
	return false
}
func (ps *ProjectScan) close() {
	ps.scn.close()
}

// PRODUCT SCAN---------------------------------
type ProductScan struct {
	s1 Scan
	s2 Scan
}

func (ps ProductScan) String() string {
	return fmt.Sprint("Product/combining (FROM ...) tables: [????] => \n")
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

func (ps *ProductScan) getInt(fldname string) (int, bool) {
	if ps.s1.hasField(fldname) {
		return ps.s1.getInt(fldname)
	} else {
		return ps.s2.getInt(fldname)
	}
	// ! throw runtime exception field not found
}

func (ps *ProductScan) getString(fldname string) (string, bool) {
	if ps.s1.hasField(fldname) {
		return ps.s1.getString(fldname)
	} else {
		return ps.s2.getString(fldname)
	}
}

func (ps *ProductScan) getVal(fldname string) (Constant, bool) {
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
