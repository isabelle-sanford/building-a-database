package main

// todo getters probably?

type QueryData struct {
	fields []string
	tables []string // collection?
	pred   Predicate
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
	predstring := qd.pred.String()
	if predstring != "" {
		result += " where " + predstring
	}
	return result
}

type InsertData struct {
	tblname string
	flds    []string
	vals    []Constant
}

type DeleteData struct {
	tblname string
	pred    Predicate
}

type ModifyData struct {
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

type CreateIndexData struct {
	idxname string
	tblname string
	fldname string
}
