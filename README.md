# About

This repository contains an implementation of a very simple database engine, written in Go. (In less technical terms, it's not just having a database of specific information. The database engine is what the data goes _into_ and stores it - kind of like building a simple version of Excel, but with more versatility and fewer handy buttons.)

This is my senior capstone project as a Computer Science major at Bryn Mawr College. I'm building this database in order to understand how the inside of a database work, from reading and writing to disk all the way to parsing SQL queries and optimizing how to get the answer most efficiently. I do a lot of work with data analysis and visualization, so I wanted to get to know the underlying architecture that I rely on so often. It may also be, I hope, a useful resource for some burgeoning database engineer out there, to modify and play with and pick apart much more easily than the databases that we generally use.

If you know anyone who would be interested in hiring a newly-graduated CS student who does data or visual or coding stuff, I would love it if you sent them [my resume](https://drive.google.com/file/d/1spyOKMx9la_qlv9Gs76IGHj2i2iiNwhj/view?usp=drive_link)!

## Status

These are the current features of the database engine. The numbers refer to the corresponding chapter in the textbook I'm using as a basis (_Database Design and Implementation_, by Edward Sciore). The same numbers are also prefixed to names of files containing the code for the corresponding feature. This means that the order of the files is in rising complexity, and later files generally build on the objects of earlier ones.

- [3] Files are divided up into blocks of a size matching what the CPU uses when writing/reading from the disk. These blocks can read and write integers, strings, and generic byte objects (blobs). Blocks are accessed independently of each other, so reading or writing one block does not require reading or writing the rest of the file.
- [3,4] To minimize disk reads and writes, currently-in-use blocks are stored in _pages_ held in a _buffer pool_, which holds pages in use but also, if there is any room left over, keeps recently-used blocks around (i.e. caches them) and does not write to the disk until required.
- [4,5] The database has a _log manager_, which records each change to the database as they occur, so that returning to a previous state is possible (whether intentionally by reverting a change, or for restoring if the database crashes). These changes are grouped into distinct _transactions_, one from each concurrent user, so that users do not interfere with each other and so that changes are not officially made until the user sends a signal to commit. (If a user crashed unexpectedly, having a partially-done set of changes could be very bad!)
- [5] Note that proper concurrency safety is _not_ implemented at the moment, and is one of the reasons that commonly used databases are so much more confusing to take apart.
- [6] Database records are stored in record pages, which can be retrieved from inside blocks. The way they are stored and accessed is determined by the _schema_ and _layout_ of each table. This is essentially what the fields (columns) of the table are called and what type they store (e.g. integer, string, blob). Accessing or modifying the record pages is done with a _table scan_.
- [7] The database keeps track of the tables via a _table manager_, which keeps lists of all tables and their fields (including field types).
- [7] The database keeps track of basic statistics about the tables, in order to significantly optimize queries.
- [8] The database has functions it can use to perform relational algebra on its data (e.g. it can filter a table based on a given criteria and return the shortened table).
- [9] The database can parse (basic) SQL queries input as strings into the relational algebra components it is familiar with, and return accurate results.

Future work:

- [5] The database is safe for concurrent users (i.e. addresses problems like two users trying to modify the same bit of data), and for recovery of the database should it be needed.
- [11,ish] The database fits with the Go [sql_driver interface](https://pkg.go.dev/database/sql/driver@go1.19.4) standard and can be used as a package like a standard Go module.
- [11,ish] The database is stored on a server and accessed from there, rather than stored directly on the client machine.
- [12] The database uses indexes and B-trees to store and access data much more efficiently.

## Resources

By far the most vital resource for me has been _Database Design and Implementation_, by Edward Sciore.
