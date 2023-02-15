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

func main() {

	db := makeDB()

	//table1 := "table1"

	createTableQuery := "create table table1 ( COL1 varchar (20) , COL2 int )"
	insertQuery := "insert into table1 ( COL1 , COL2 ) values ( 'Hello' , 20 )"

	tx := db.makeTx()

	db.planner.executeUpdate(createTableQuery, tx)

	// db.mdm.tm.showTblCatalog()
	// db.mdm.tm.showFldCatalog()

	fmt.Println("table1 layout:", db.mdm.getLayout("table1", tx))

	db.planner.executeUpdate(insertQuery, tx)

	// ! todo - make 'printTable(tblname)' method in table manager
	// DEFINITELY DO NOT DO THIS WTF
	table1 := makeTableScan(tx, "table1", db.mdm.getLayout("table1", tx))
	table1.printTable()

	db.mdm.sm.refreshStatistics(tx)

	si := db.mdm.getStatInfo("table1", db.mdm.getLayout("table1", tx), tx)
	fmt.Println("\nstat info:", si)

	fmt.Println("statmgr table stats:", db.mdm.sm.tableStats)

	tx.commit()

}
