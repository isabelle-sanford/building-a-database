package main

import "fmt"

type TableMgr struct {
	tx     *Transaction
	tblcat Layout
	fldcat Layout // ??
}

type tblInfo struct {
	tblname  string
	slotsize int
}

type fldInfo struct {
	tblname string
	fldname string
	fldtype int
	length  int
	offset  int
}

const MAX_NAME_LEN = 20

// ...actually maybe just make these as schema?

func makeTableMgr(tx *Transaction, isnew bool) TableMgr {
	tblcat := makeTblCat()
	fldcat := makeFldCat()

	tm := TableMgr{tx, tblcat, fldcat}

	if isnew {
		// make catalog tables if db is new
		tm.createTable("tblcat", tblcat.schema, tx)
		tm.createTable("fldcat", fldcat.schema, tx)
	}

	return tm
}

func (tm *TableMgr) createTable(tblname string, sch Schema, tx *Transaction) {
	l := makeLayoutFromSchema(sch)

	tcat := makeTableScan(tx, "tblcat", tm.tblcat) // ! pull from tm name at least
	tcat.insert()
	tcat.setString("tblname", tblname)
	tcat.setInt("slotsize", l.slotsize)
	tcat.close()

	fcat := makeTableScan(tx, "fldcat", tm.fldcat)
	for fldname, fldinfo := range sch.fields {
		fcat.insert()
		fcat.setString("tblname", tblname)
		fcat.setString("fldname", fldname)
		fcat.setInt("fldtype", fldinfo.fldtype)
		fcat.setInt("length", fldinfo.length)
		fcat.setInt("offset", l.offsets[fldname])
	}
	fcat.close()
}

func (tm *TableMgr) getLayout(tblname string, tx *Transaction) Layout {
	var sch Schema = makeSchema()

	fcat := makeTableScan(tx, "fldcat", tm.tblcat)

	rec := fcat.next()
	for rec {
		tbl := fcat.getString("tblname")
		if tbl == tblname { // this is ok right?
			fldname := fcat.getString("fldname")
			fldtype := fcat.getInt("fldtype")
			fldlen := fcat.getInt("length")
			sch.addField(fldname, fldtype, fldlen)
		}
	}
	// could also pull this during above info
	l := makeLayoutFromSchema(sch)
	fcat.close()
	return l
}

func makeTblCat() Layout {
	var sch Schema = makeSchema()
	sch.addStringField("tblname", MAX_NAME_LEN)
	sch.addIntField("slotsize")
	l := makeLayoutFromSchema(sch)
	return l
}

func makeFldCat() Layout {
	var sch Schema = makeSchema()
	sch.addStringField("tblname", MAX_NAME_LEN)
	sch.addStringField("fldname", MAX_NAME_LEN)
	sch.addIntField("fldtype")
	sch.addIntField("length")
	sch.addIntField("offset")
	l := makeLayoutFromSchema(sch)
	return l
}

func main() {
	db := makeDB()
	tx := db.makeTx()
	tm := makeTableMgr(tx, true)

	var sch Schema = makeSchema()
	sch.addIntField("A")
	sch.addStringField("B", 9)

	tm.createTable("MyTable", sch, tx)
	layout := tm.getLayout("MyTable", tx)

	sch2 := layout.schema
	fmt.Println("MyTable has slot size ", layout.slotsize)
	fmt.Println("Its fields are:")
	for fldname, fldinfo := range sch2.fields {
		var fldtype string
		if fldinfo.fldtype == INTEGER {
			fldtype = "int"
		} else {
			strlen := fldinfo.length
			fldtype = fmt.Sprint("varchar(", strlen, ")")
		}
		fmt.Println(fldname, ": ", fldtype)
	}
	tx.commit()
}
