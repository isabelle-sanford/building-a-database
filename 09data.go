package main

// todo getters probably?

type QueryData struct {
	fields []string
	tables []string // collection?
	pred   *Predicate
}

func (qd QueryData) String() string {
	result := "select "
	for _, fldname := range qd.fields {
		result += fldname + ", "
	}
	// todo zap final comma
	result += " from "
	for _, tblname := range qd.tables {
		result += tblname + ", "
	}
	// todo zap final comma
	if qd.pred != nil {
		predstring := qd.pred.String()
		if predstring != "" {
			result += " where " + predstring
		}
	}

	return result
}

// ?? i guess
type UpdateData interface{}

type InsertData struct { // tableName, fields, vals
	tblname string
	flds    []string
	vals    []Constant
}

type DeleteData struct { // tableName, pred
	tblname string
	pred    Predicate
}

type ModifyData struct { // tableName, targetField, newValue, pred
	tblname string
	fldname string
	newval  Expression
	pred    Predicate
}

type CreateTableData struct {
	tblname string
	sch     Schema
}

type CreateViewData struct {
	viewname string
	qrdata   QueryData
}

// could make all getters funcs like this?
func (cvw CreateViewData) viewDef() string {
	return cvw.qrdata.String()
}

type CreateIndexData struct {
	idxname string
	tblname string
	fldname string
}
