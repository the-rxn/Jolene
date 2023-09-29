package myDb

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	sql    *sql.DB
	stmt   *sql.Stmt
	buffer []StorageLine
}
type StorageLine struct {
	UserID  int64
	Message string
	Time    time.Time
}

const (
	insertSQL = `
	INSERT INTO MAIN 
	(userid, message, time) 
	VALUES (?, ?, ?)
	`
	schemaSQL = `
	CREATE TABLE IF NOT EXISTS main (
	userID INTEGER,
	message STRING,
	time TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS main_time ON main(time);
	CREATE INDEX IF NOT EXISTS main_userID ON main(userID);

	`
)

func NewDB(dbFile string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, err
	}
	db := DB{
		sql:    sqlDB,
		stmt:   stmt,
		buffer: make([]StorageLine, 0, 512),
	}
	return &db, nil
}
func (db *DB) Add(storageLine StorageLine) error {
	if len(db.buffer) == cap(db.buffer) {
		return errors.New("storageLines buffer is full")
	}

	db.buffer = append(db.buffer, storageLine)
	if len(db.buffer) == cap(db.buffer) {
		if err := db.Flush(); err != nil {
			log.Error("unable to flush storageLines: %w", err)
			return err
		}
	}

	return nil
}

// Flush pending txs into DB.
func (db *DB) Flush() error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}

	for _, storageLine := range db.buffer {
		_, err := tx.Stmt(db.stmt).Exec(storageLine.UserID, storageLine.Message, storageLine.Time)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	db.buffer = db.buffer[:0]
	return tx.Commit()
}

func (db *DB) Close() error {
	defer func() {
		db.stmt.Close()
		db.sql.Close()
	}()

	if err := db.Flush(); err != nil {
		return err
	}

	return nil
}
