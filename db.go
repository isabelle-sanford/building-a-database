package main

import (
	"fmt"
	"log"
	"os"
)

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

func testCreate(tblname string, tx *Transaction, db myDB, logger *log.Logger, col1 string, col2 string) {
	createTableQuery := fmt.Sprintf("create table %s ( %s varchar (20) , %s int )", tblname, col1, col2)

	logger.Print("Executing query ", createTableQuery)

	db.planner.executeUpdate(createTableQuery, tx)

	logger.Print("Table created!")
	logger.Print(db.mdm.getLayout(tblname, tx))
	logger.Print("Catalogs - ")
	db.mdm.tm.showTblCatalog(logger)
	db.mdm.tm.showFldCatalog(logger)

}

func testInsert(tblname string, tx *Transaction, db myDB, logger *log.Logger) {

	insertQuery := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'Hello' , 20 )"
	insert2 := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'World' , 10 )"
	insert3 := "insert into " + tblname + " ( COL1 , COL2 ) values ( 'Diamond' , 20 )"

	// INSERTIONS
	db.planner.executeUpdate(insertQuery, tx)
	db.planner.executeUpdate(insert2, tx)
	db.planner.executeUpdate(insert3, tx)

	logger.Print("Insertions complete! Table now looks like")
	logger.Print(db.mdm.tm.printTable(tblname, tx))

	db.mdm.sm.refreshStatistics(tx)

	logger.Print(db.mdm.getStatInfo(tblname, db.mdm.getLayout(tblname, tx), tx))

}

func doInsert(tblname string, tx *Transaction, db myDB, logger *log.Logger, val1 string, val2 int, col1 string, col2 string) {
	insertQuery := fmt.Sprintf("insert into %s ( %s , %s ) values ( '%s' , %d )", tblname, col1, col2, val1, val2)

	db.planner.executeUpdate(insertQuery, tx)

	logger.Print("Insertions complete! Table now looks like")
	logger.Print(db.mdm.tm.printTable(tblname, tx))

	db.mdm.sm.refreshStatistics(tx)

	logger.Print(db.mdm.getStatInfo(tblname, db.mdm.getLayout(tblname, tx), tx))

}

func testQuery(tblname string, tx *Transaction, db myDB, logger *log.Logger, query string) {
	//projectQuery := "select COL1 from table1 "
	p := db.planner.createQueryPlan(query, tx)

	logger.Print("\nQuery complete! ")
	logger.Print("\nPLAN: \n", p)

	logger.Print("SCAN (i.e. results):")
	printResult(p, logger)

}

func printResult(p Plan, logger *log.Logger) {
	ret := ""
	scn := p.open()
	schem := p.schema()
	scn.beforeFirst()

	for scn.next() {
		ret += stringScanRecord(*schem, scn) + "\n"
	}

	logger.Print(ret)

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

	logfile, _ := os.Create("dblogs.txt")
	logger := log.New(logfile, "logger: ", log.Lshortfile)

	log2, _ := os.Create("output.txt")
	logger2 := log.New(log2, " ", 0)

	db := makeDB()

	tblname := "table1"
	tbl2 := "tbl2"

	tx := db.makeTx()

	testCreate(tblname, tx, db, logger, "col1", "col2")
	testCreate(tbl2, tx, db, logger, "col4", "col3")

	doInsert(tblname, tx, db, logger, "hello", 20, "col1", "col2")
	doInsert(tblname, tx, db, logger, "world", 10, "col1", "col2")
	doInsert(tblname, tx, db, logger, "yoyo", 12, "col1", "col2")
	doInsert(tblname, tx, db, logger, "hello", 10, "col1", "col2")
	doInsert(tbl2, tx, db, logger, "diamond", 100, "col4", "col3")
	doInsert(tbl2, tx, db, logger, "love", 13, "col4", "col3")
	doInsert(tbl2, tx, db, logger, "fdasf", 23, "col4", "col3")
	doInsert(tbl2, tx, db, logger, "hello hello", 23, "col4", "col3")

	query := "select col1, col2, col3, col4  from table1, tbl2 where col2 = 20"
	testQuery(tblname, tx, db, logger2, query)

	tx.commit()

}
