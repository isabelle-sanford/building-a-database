package main

type myDB struct {
	fm *FileMgr
	lm *LogMgr
	bm *BufferMgr
}

func makeDB() myDB {

	fm := makeFileMgr("mydb", 400)

	lm := makeLogMgr(&fm, "log.log")

	bm := makeBufferManager(&fm, &lm, 9)

	return myDB{&fm, &lm, &bm}
}

func (db *myDB) makeTx() *Transaction {
	return makeTransaction(db.fm, db.lm, db.bm)
}
