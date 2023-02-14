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
	//fmt.Println("openFiles at start of tblmgr constructor: ", tx.fm.openFiles)

	// layouts
	tblcat := makeTblCat()
	fldcat := makeFldCat()

	//fmt.Println("openFiles in tblmgr constructor after makeTblCat and fldCat: ", tx.fm.openFiles)

	tm := TableMgr{tx, tblcat, fldcat}

	//fmt.Println("Made tblcat and fldcat. Creating if new...")

	if isnew {
		// make catalog tables if db is new
		tm.createTable("tblcat", tblcat.schema, tx)
		tm.createTable("fldcat", fldcat.schema, tx)
	}

	return tm
}

func (tm *TableMgr) createTable(tblname string, sch Schema, tx *Transaction) {
	l := makeLayoutFromSchema(sch)

	// insert record of table into table catalog
	tcat := makeTableScan(tx, "tblcat", tm.tblcat) // ! pull from tm name at least
	tcat.insert()
	tcat.setString("tblname", tblname)
	tcat.setInt("slotsize", l.slotsize)
	tcat.close()

	// insert record of all fields into field catalog
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

	// open field catalog
	fcat := makeTableScan(tx, "fldcat", tm.fldcat)

	for fcat.next() {
		// tblname of row
		tbl := fcat.getString("tblname")

		// if this field is in the table we're finding
		if tbl == tblname { // this is ok right?
			fldname := fcat.getString("fldname")
			fldtype := fcat.getInt("fldtype")
			fldlen := fcat.getInt("length")
			sch.addField(fldname, fldtype, fldlen)
		}
	}

	//fmt.Println(sch)

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

func (tm *TableMgr) showTblCatalog() {

	tcat := makeTableScan(tm.tx, "tblcat", tm.tblcat)

	fmt.Println("Table Catalog: ")
	tcat.printTable()
	fmt.Println()

	tcat.close()
}

func (tm *TableMgr) showFldCatalog() {
	fcat := makeTableScan(tm.tx, "fldcat", tm.fldcat)

	fmt.Println("Field Catalog: ")
	fcat.printTable()
	fmt.Println()
	fcat.close()
}

func CatalogTest() {
	db := makeDB()
	tx := db.makeTx()
	tm := makeTableMgr(tx, true)

	fmt.Println("Catalog test:")
	var sch Schema = makeSchema()
	sch.addIntField("A")
	sch.addStringField("B", 9)

	fmt.Println("Creating table 'MyTable'...")
	tm.createTable("MyTable", sch, tx)

	tm.showTblCatalog()
	tm.showFldCatalog()

	fmt.Println()

	tCat := makeTableScan(tx, "tblcat", tm.tblcat)

	fmt.Println("\nAll tables and their lengths:")
	for tCat.next() {
		tname := tCat.getString("tblname")
		tsize := tCat.getInt("slotsize")
		fmt.Println(tname, " ", tsize)
	}
	tCat.close()
	fmt.Println()

	fCat := makeTableScan(tx, "fldcat", tm.fldcat)

	fmt.Println("\nAll fields and their offsets:")
	for fCat.next() {

		tname := fCat.getString("tblname")
		fname := fCat.getString("fldname")
		offset := fCat.getInt("offset")
		fmt.Printf("Table: %s Field: %s Offset: %d \n", tname, fname, offset)
		//fmt.Println(tname, " ", fname, " ", offset)
	}
	fCat.close()
}

func tableMgrTest() {
	db := makeDB()
	tx := db.makeTx()
	tm := makeTableMgr(tx, true)

	fmt.Println("Catalog test:")
	var sch Schema = makeSchema()
	sch.addIntField("A")
	sch.addStringField("B", 9)

	fmt.Println("Creating table 'MyTable'...")
	tm.createTable("MyTable", sch, tx)

	fmt.Println("\nTM test:")
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
