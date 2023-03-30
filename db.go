package main

import "fmt"

type myDB struct {
	fm      *FileMgr
	lm      *LogMgr
	bm      *BufferMgr
	mdm     *MetadataMgr
	planner *Planner
}

func makeDB() myDB {

	fm := makeFileMgr("mydb", 400)

	lm := makeLogMgr(&fm, "log.log")

	bm := makeBufferManager(&fm, &lm, 9)

	tx := makeTransaction(&fm, &lm, &bm) // idk

	// maybe this should just return a *MetadataManager
	mdm := makeMetadataMgr(true, tx) // contains table and stat mgrs

	// change these for different planning stuffs
	qp := BasicQueryPlanner{*mdm}
	up := BasicUpdatePlanner{*mdm}

	planner := Planner{qp, up}

	return myDB{&fm, &lm, &bm, mdm, &planner}
}

func (db *myDB) makeTx() *Transaction {
	return makeTransaction(db.fm, db.lm, db.bm)
}

func testCreate(tblname string, tx *Transaction, db myDB, prints bool) {
	createTableQuery := "create table " + tblname + " ( COL1 varchar (20) , COL2 int )"

	if prints {
		fmt.Println("Executing query ", createTableQuery)
	}

	db.planner.executeUpdate(createTableQuery, tx)

	if prints {
		fmt.Println("Table created!")
		fmt.Println(db.mdm.getLayout(tblname, tx))

		fmt.Println("Catalogs - ")
		db.mdm.tm.showTblCatalog()
		db.mdm.tm.showFldCatalog()
	}
}

func testInsert(tblname string, tx *Transaction, db myDB, prints bool) {

	insertQuery := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'Hello' , 20 )"
	insert2 := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'World' , 10 )"
	insert3 := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'Diamond' , 20 )"

	// INSERTIONS
	db.planner.executeUpdate(insertQuery, tx)
	db.planner.executeUpdate(insert2, tx)
	db.planner.executeUpdate(insert3, tx)

	if prints {
		fmt.Println("Insertions complete! Table now looks like")
		db.mdm.tm.printTable(tblname, tx)
	}

	db.mdm.sm.refreshStatistics(tx)

	if prints {
		fmt.Println(db.mdm.getStatInfo(tblname, db.mdm.getLayout(tblname, tx), tx))
	}
}

func testQuery(tblname string, tx *Transaction, db myDB, prints bool) {
	//projectQuery := "select COL1 from table1 "

	selectQuery := "select COL1, COL2 from table1 where COL2 = 20"
	p := db.planner.createQueryPlan(selectQuery, tx)

	if prints {
		fmt.Println("\nQuery complete! ")
		fmt.Println("\nPLAN: \n", p)
	}

	if prints {
		fmt.Println("SCAN (i.e. results):")
		printResult(p)
	}
}

func printResult(p Plan) {
	ret := ""
	scn := p.open()
	schem := p.schema()
	scn.beforeFirst()

	for scn.next() {
		ret += stringScanRecord(*schem, scn) + "\n"
	}

	fmt.Println(ret)

	scn.close()
}

func stringScanRecord(sch Schema, scn Scan) string {
	var s string
	var val string
	for _, fldname := range sch.fieldlist {
		fld := sch.fields[fldname]
		if fld.fldtype == VARCHAR {
			val, _ = scn.getString(fldname)
		} else {
			temp, _ := scn.getInt(fldname)
			val = fmt.Sprint(temp)
		}

		s += val + " "
	}
	return s
}

func main() {

	db := makeDB()

	tblname := "table1"
	tbl2 := "tbl2"

	tx := db.makeTx()

	testCreate(tblname, tx, db, true)

	testInsert(tblname, tx, db, true)

	testQuery(tblname, tx, db, true)

	tx.commit()

}
