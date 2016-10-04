package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// Storage interface provides methods to use for other code of app, so it doesn't depend on storage implementation.
type Storage interface {
	Put(string, []byte) error
	Get(string) ([]byte, error)
	Remove(string) error
	Next() ([]byte, error)
}

// SqliteStorage embedded storage.
type SqliteStorage struct {
	dbPath string
}

// Init initialise sqlite storage.
func (storage *SqliteStorage) Init() error {
	// Create sqlite db file
	db, err := sql.Open("sqlite3", storage.dbPath)
	if err != nil {
		return fmt.Errorf("Storage: unable to create db. %s", err.Error())
	}
	defer db.Close()
	// Create table if it' doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS jobs(
		id TEXT NOT NULL PRIMARY KEY,
		addedAt DATETIME,
		data BLOB
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("Storage: unable to create jobs table. %s", err.Error())
	}
	return nil
}

// Put saves job to the storage.
func (storage *SqliteStorage) Put(id string, data []byte) error {
	db, err := sql.Open("sqlite3", storage.dbPath)
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, open db failed. %s", id, err.Error())
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, create transaction failed. %s", id, err.Error())
	}
	stmt, err := tx.Prepare("INSERT INTO jobs(id, addedAt, data) VALUES(?, datetime('now'), ?)")
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, unable to prepare statement. %s", id, err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, data)
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, execute failed. %s", id, err.Error())
	}
	tx.Commit()

	return nil
}

// Get saves job to the storage.
func (storage *SqliteStorage) Get(id string) ([]byte, error) {
	db, err := sql.Open("sqlite3", storage.dbPath)
	if err != nil {
		return nil, fmt.Errorf("Storage: GET %s, open db failed. %s", id, err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT data FROM jobs WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("Storage: GET %s, prepare statement failed. %s", id, err.Error())
	}
	defer stmt.Close()

	var data []byte
	row := stmt.QueryRow(id)
	if row == nil {
		return nil, nil
	}
	row.Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("Storage: GET %s, prepare statement failed. %s", id, err.Error())
	}

	return data, nil
}

// Remove job from the storage by id
func (storage *SqliteStorage) Remove(id string) error {
	db, err := sql.Open("sqlite3", storage.dbPath)
	if err != nil {
		return fmt.Errorf("Storage: REMOVE %s, open db failed. %s", id, err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM jobs WHERE id = ?")
	if err != nil {
		return fmt.Errorf("Storage: REMOVE %s, prepare statement failed. %s", id, err.Error())
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("Storage: REMOVE %s, delete query failed. %s", id, err.Error())
	}

	return nil
}

// Next returns any item needs to be processed.
func (storage *SqliteStorage) Next() ([]byte, error) {
	return nil, nil
}

// NewStorage create new storage entity.
func NewStorage(dbPath string) (Storage, error) {

	if dbPath == "" {
		return nil, fmt.Errorf("Storage: please provide non-empty path to the storage")
	}

	storage := SqliteStorage{dbPath: dbPath}
	err := storage.Init()
	if err != nil {
		return nil, err
	}

	return &storage, nil
}
