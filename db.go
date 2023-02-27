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

func testInsert(tblname string, tx *Transaction, db myDB) {
	createTableQuery := "create table table1 ( COL1 varchar (20) , COL2 int )"
	insertQuery := "insert into table1 ( COL1 , COL2 ) values ( 'Hello' , 20 )"
	insert2 := "insert into table1 ( COL1 , COL2 ) values ( 'World' , 10 )"
	insert3 := "insert into table1 ( COL1 , COL2 ) values ( 'Diamond' , 15 )"

	// CREATE TABLE
	db.planner.executeUpdate(createTableQuery, tx)
	// db.mdm.tm.showTblCatalog()
	// db.mdm.tm.showFldCatalog()
	//fmt.Println("table layout:", db.mdm.getLayout(tblname, tx))

	// INSERTIONS
	db.planner.executeUpdate(insertQuery, tx)
	//db.mdm.tm.printTable(tblname, tx)

	//db.mdm.sm.refreshStatistics(tx)
	//si := db.mdm.getStatInfo(tblname, db.mdm.getLayout(tblname, tx), tx)
	//fmt.Println("\nstat info:", si)

	//fmt.Println("statmgr table stats:", db.mdm.sm.tableStats)

	db.planner.executeUpdate(insert2, tx)
	db.planner.executeUpdate(insert3, tx)

	fmt.Println("Created table and inserted 3 records. Table is now: ")
	db.mdm.tm.printTable(tblname, tx)
}

func main() {

	db := makeDB()

	tblname := "table1"

	tx := db.makeTx()

	testInsert(tblname, tx, db)

	// QUERIES
	fmt.Println("Querying!")

	projectQuery := "select COL1 from table1 where "

	//selectQuery := "select COL1, COL2 from table1 where COL2 = 20"

	db.planner.createQueryPlan(projectQuery, tx)

	//fmt.Println(p)

	tx.commit()

}
