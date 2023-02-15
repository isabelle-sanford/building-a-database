package main

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

	tx := db.makeTx()

	db.planner.executeUpdate(createTableQuery, tx)

}
