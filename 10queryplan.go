package main

import "fmt"

type QueryPlanner interface {
	createPlan(data QueryData, tx *Transaction) Plan
}

type UpdatePlanner interface {
	executeInsert(data InsertData, tx *Transaction) int
	executeDelete(data DeleteData, tx *Transaction) int
	executeModify(data ModifyData, tx *Transaction) int
	executeCreateTable(data CreateTableData, tx *Transaction) int
	executeCreateView(data CreateViewData, tx *Transaction) int
	executeCreateIndex(data CreateIndexData, tx *Transaction) int
}

type BasicQueryPlanner struct {
	mdm MetadataMgr
}

type BasicUpdatePlanner struct {
	mdm MetadataMgr
}

type Planner struct {
	qp QueryPlanner
	up UpdatePlanner
}

// PLANNER
func (pl Planner) createQueryPlan(cmd string, tx *Transaction) Plan {
	parser := makeParser(cmd)
	data := parser.query()
	// code to verify query should go here

	fmt.Println("Parsed query: ", data)

	return pl.qp.createPlan(data, tx)
}
func (pl Planner) executeUpdate(cmd string, tx *Transaction) int {
	parser := makeParser(cmd)

	// fmt.Println("\nparser pre-parse:\n", parser.lex)

	obj, typ := parser.updateCmd() // HMPH
	// code to verify update cmd goes here

	//fmt.Println("\nupdate is", obj, typ)
	// fmt.Println("\nparser post-update:\n", parser.lex)

	switch typ {
	case "insert":
		return pl.up.executeInsert(obj.(InsertData), tx)
	case "delete":
		return pl.up.executeDelete(obj.(DeleteData), tx)
	case "modify":
		return pl.up.executeModify(obj.(ModifyData), tx)
	case "createTable":
		return pl.up.executeCreateTable(obj.(CreateTableData), tx)
	case "createView":
		return pl.up.executeCreateView(obj.(CreateViewData), tx)
	case "createIndex":
		return pl.up.executeCreateIndex(obj.(CreateIndexData), tx)
	default:
		return 0 // should never happen
	}

}

// QUERY PLANNER
// not pointering is ok right?
func (bqp BasicQueryPlanner) createPlan(data QueryData, tx *Transaction) Plan {
	plans := make([]Plan, 0)
	for _, tblname := range data.tables {
		// SHOULD check whether it's a view (mdm.getViewDef(tblname, tx))
		// and do other stuff if so

		// else make table plan for plans list
		plans = append(plans, makeTablePlan(tx, tblname, bqp.mdm))
	}

	fmt.Println("Plans from table list: ", plans)

	// product all plans together (if there's more than 1)
	p := plans[0]
	if len(plans) > 1 {
		for _, nextplan := range plans[1:] {

			p1 := makeProductPlan(p, nextplan)
			p2 := makeProductPlan(nextplan, p)
			if p1.blocksAccessed() < p2.blocksAccessed() {
				p = p1
			} else {
				p = p2
			}
		}
	}

	p = SelectPlan{p, *data.pred}

	return makeProjectPlan(p, data.fields)
}

// UPDATE PLANNER
func (bup BasicUpdatePlanner) executeDelete(data DeleteData, tx *Transaction) int {
	p := makeTablePlan(tx, data.tblname, bup.mdm)
	s := SelectPlan{p, data.pred} // can't just cast p to selectscan sadly
	us := s.open().(UpdateScan)
	count := 0

	for us.next() {
		us.delete() // hmph
		count++
	}
	us.close()
	return count
}
func (bup BasicUpdatePlanner) executeModify(data ModifyData, tx *Transaction) int {
	p := makeTablePlan(tx, data.tblname, bup.mdm)
	s := SelectPlan{p, data.pred} // can't just cast p to selectscan sadly
	us := s.open().(UpdateScan)
	count := 0

	for us.next() {
		val := data.newval.evaluate(us)
		us.setVal(data.fldname, val)
	}
	us.close()
	return count
}
func (bup BasicUpdatePlanner) executeInsert(data InsertData, tx *Transaction) int {
	p := makeTablePlan(tx, data.tblname, bup.mdm)
	us := p.open().(UpdateScan)

	us.insert()

	count := 0 // ! crude replacement for iterator

	for _, fldname := range data.flds {
		val := data.vals[count] // ew
		us.setVal(fldname, val)
		count++
	}
	us.close()
	return 1
}
func (bup BasicUpdatePlanner) executeCreateTable(data CreateTableData, tx *Transaction) int {
	bup.mdm.createTable(data.tblname, data.sch, tx)
	return 0
}
func (bup BasicUpdatePlanner) executeCreateView(data CreateViewData, tx *Transaction) int {
	bup.mdm.createView(data.viewname, data.viewDef(), tx)
	return 0
}
func (bup BasicUpdatePlanner) executeCreateIndex(data CreateIndexData, tx *Transaction) int {
	bup.mdm.createIndex(data.idxname, data.tblname, data.fldname, tx)
	return 0
}
