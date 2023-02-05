package main

import "math"

type Constant struct {
	ival  int
	sval  string
	isInt bool
}

type Expression struct {
	val     Constant
	fldname string
	isConst bool
}

type Term struct {
	lhs, rhs Expression
}

type Predicate struct {
	terms []Term
}

const INT_MAX_VAL = 90000000 // !!

// CONSTANT
func makeConstInt(ival int) Constant {
	return Constant{ival: ival}
}

func makeConstString(sval string) Constant {
	return Constant{sval: sval} // this works right?
}

// might not be necessary
func (c Constant) asInt() int {
	return c.ival
}
func (c Constant) asString() string {
	return c.sval
}

// equals func ?
func (c Constant) equals(c1 Constant) bool {
	if c.isInt {
		return c.ival == c1.ival
	} else {
		return c.sval == c1.sval // is this fine in go?
	}
}

// compare func ?
func (c Constant) compare(c1 Constant) int {
	if c.isInt {
		return c.ival - c1.ival
	} else {
		return c.sval.compare(c1.sval) // HOW TO COMPARE STRINGS
	}
}

// hashcode func ?

func (c Constant) String() string {
	if c.isInt {
		return string(c.ival)
	} else {
		return c.sval
	}
}

// EXPRESSION

func makeExprConst(val Constant) Expression {
	return Expression{val: val, isConst: true}
}
func makeExprFld(fldname string) Expression {
	return Expression{fldname: fldname}
}
func (e Expression) isFieldName() bool {
	return e.fldname != "" // ? does this work?
}

// getters?
func (e Expression) evaluate(s Scan) Constant {
	if e.isConst {
		return e.val
	} else {
		return s.getVal(e.fldname)
	}
}
func (e Expression) appliesTo(sch Schema) bool {
	if e.isConst {
		return true
	} else {
		return sch.hasField(e.fldname)
	}
}
func (e Expression) String() string {
	if e.isConst {
		return e.val.String()
	} else {
		return e.fldname
	}
}

// TERM
func (t Term) isSatisfied(s Scan) bool {
	lhsval := t.lhs.evaluate(s)
	rhsval := t.rhs.evaluate(s)
	return rhsval.equals(lhsval)
}
func (t Term) appliesTo(sch Schema) bool {
	return t.lhs.appliesTo(sch) && t.rhs.appliesTo(sch)
}
func (t Term) reductionFactor(p Plan) int {
	var lhsName, rhsName string
	if t.lhs.isFieldName() && t.rhs.isFieldName() {
		lhsName = t.lhs.fldname
		rhsName = t.rhs.fldname
		// !
		return int(math.Max(p.distinctValues(lhsName), p.distinctValues(rhsName)))
	}
	if t.lhs.isFieldName() {
		lhsName = t.lhs.fldname
		return p.distinctValues(lhsName)
	}
	if t.rhs.isFieldName() {
		rhsName = t.rhs.fldname
		return p.distinctValues(rhsName)
	}

	// otherwise, the term equates constants
	if t.lhs.val.equals(t.rhs.val) {
		return 1
	} else {
		return INT_MAX_VAL
	}
}
func (t Term) equatesWithConstant(fldname string) *Constant {
	if t.lhs.isFieldName() && t.lhs.fldname == fldname && !t.rhs.isFieldName() {
		return &t.rhs.val
	} else if t.rhs.isFieldName() && t.rhs.fldname == fldname && !t.lhs.isFieldName() {
		return &t.lhs.val
	} else {
		return nil // ????? should be returning nil
	}
}
func (t Term) equatesWithField(fldname string) *string {
	if t.lhs.isFieldName() && t.lhs.fldname == fldname && !t.rhs.isFieldName() {
		return &t.rhs.fldname
	} else if t.rhs.isFieldName() && t.rhs.fldname == fldname && !t.lhs.isFieldName() {
		return &t.lhs.fldname
	} else {
		return nil // ????? should be returning nil
	}
}
func (t Term) String() string {
	return t.lhs.String() + "=" + t.rhs.String()
}

// PREDICATE
func makePredwTerm(t Term) Predicate {
	terms := make([]Term, 1)
	terms[0] = t
	return Predicate{terms}
}

func (pred *Predicate) conjoinWith(pred1 *Predicate) {
	pred.terms = append(pred.terms, pred1.terms...)
}

func (pred *Predicate) isSatisfied(s Scan) bool {
	for _, t := range pred.terms {
		if !t.isSatisfied(s) {
			return false
		}
	}
	return true
}

func (pred *Predicate) reductionFactor(p Plan) int {
	factor := 1
	for _, t := range pred.terms {
		factor *= t.reductionFactor(p)
	}
	return factor
}

func (pred *Predicate) selectSubPred(sch Schema) *Predicate {
	var result Predicate
	for _, t := range pred.terms {
		if t.appliesTo(sch) {
			result.terms = append(result.terms, t)
		}
	}
	if len(result.terms) == 0 {
		return nil // !!
	} else {
		return &result
	}
}

func (pred *Predicate) joinSubPred(sch1 Schema, sch2 Schema) *Predicate {
	var result Predicate
	var newsch Schema
	newsch.addAll(sch1)
	newsch.addAll(sch2)
	for _, t := range pred.terms {
		if !t.appliesTo(sch1) && !t.appliesTo(sch2) && t.appliesTo(newsch) {
			result.terms = append(result.terms, t)
		}
	}
	if len(result.terms) == 0 {
		return nil // !!
	} else {
		return &result
	}
}

func (pred *Predicate) equatesWithConstant(fldname string) *Constant {
	for _, t := range pred.terms {
		c := t.equatesWithConstant(fldname)
		if &c != nil {
			return c
		}
	}
	return nil
}

func (pred *Predicate) equatesWithField(fldname string) *string {
	for _, t := range pred.terms {
		s := t.equatesWithField(fldname)
		if s != nil {
			return s
		}
	}
	return nil
}

func (pred *Predicate) String() string {

}
