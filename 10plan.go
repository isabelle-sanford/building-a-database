package main

import "math"

type Plan interface {
	open() Scan
	blocksAccessed() int // probs unnecessary tbh but eh
	recordsOutput() int
	distinctValues(fldname string) int
	schema() *Schema
}

type TablePlan struct {
	tx      *Transaction
	tblname string
	layout  Layout
	si      *StatInfo
}

type SelectPlan struct {
	p    Plan
	pred Predicate
}

type ProjectPlan struct {
	p   Plan
	sch *Schema
}

type ProductPlan struct {
	p1, p2 Plan
	sch    *Schema
}

func myMin(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func makeTablePlan(tx *Transaction, tblname string, md MetadataMgr) TablePlan {
	layout := md.getLayout(tblname, tx)
	si := md.getStatInfo(tblname, layout, tx)
	return TablePlan{tx, tblname, layout, si}
}

func (tp TablePlan) open() Scan {
	t := makeTableScan(tp.tx, tp.tblname, tp.layout)
	return &t
}
func (tp TablePlan) blocksAccessed() int {
	return tp.si.numblocks
}
func (tp TablePlan) recordsOutput() int {
	return tp.si.numrecs
}
func (tp TablePlan) distinctValues(fldname string) int {
	return tp.si.getDistinct(fldname)
}
func (tp TablePlan) schema() *Schema {
	return &tp.layout.schema
}

// SELECT SCAN
// don't need a constructor
func (sp SelectPlan) open() Scan {
	s := sp.p.open()
	return &SelectScan{s, sp.pred}
}
func (sp SelectPlan) blocksAccessed() int {
	return sp.p.blocksAccessed() // sigh // could we not just anonymize p
}
func (sp SelectPlan) recordsOutput() int {
	return sp.p.recordsOutput()
}
func (sp SelectPlan) distinctValues(fldname string) int {
	if sp.pred.equatesWithConstant(fldname) != nil {
		return 1
	} else {
		fldname2 := sp.pred.equatesWithField(fldname) // why does this return a *string tho
		if fldname2 != nil {
			return myMin(sp.p.distinctValues(fldname), sp.p.distinctValues(*fldname2))
		} else {
			return sp.p.distinctValues(fldname)
		}
	}
}
func (sp SelectPlan) schema() *Schema {
	return sp.p.schema()
}

func makeProjectPlan(p Plan, fieldlist []string) ProjectPlan {
	s := makeSchema()
	for _, fldname := range fieldlist {
		s.add(fldname, *p.schema())
	}
	return ProjectPlan{p, &s}
}

func (pp ProjectPlan) open() Scan {
	s := pp.open()
	return &ProjectScan{s, pp.sch.fieldlist}
}
func (pp ProjectPlan) blocksAccessed() int {
	return pp.p.blocksAccessed()
}
func (pp ProjectPlan) recordsOutput() int {
	return pp.p.recordsOutput()
}
func (pp ProjectPlan) distinctValues(fldname string) int {
	return pp.p.distinctValues(fldname)
}
func (pp ProjectPlan) schema() *Schema {
	return pp.sch
}

func makeProductPlan(p1, p2 Plan) ProductPlan {
	s := makeSchema()
	s.addAll(*p1.schema())
	s.addAll(*p2.schema())
	return ProductPlan{p1, p2, &s}
}

func (pop ProductPlan) open() Scan {
	s1 := pop.p1.open()
	s2 := pop.p2.open()
	return &ProductScan{s1, s2}
}
func (pop ProductPlan) blocksAccessed() int {
	return pop.p1.blocksAccessed() + pop.p1.recordsOutput()*pop.p2.blocksAccessed()
}
func (pop ProductPlan) recordsOutput() int {
	return pop.p1.recordsOutput() * pop.p2.recordsOutput()
}
func (pop ProductPlan) distinctValues(fldname string) int {
	if pop.p1.schema().hasField(fldname) {
		return pop.p1.distinctValues(fldname)
	} else {
		return pop.p2.distinctValues(fldname)
	}
}
func (pop ProductPlan) schema() *Schema {
	return pop.sch
}
