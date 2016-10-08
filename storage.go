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
	ReadAll() ([][]byte, error)
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
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO jobs(id, addedAt, data) VALUES(?, datetime('now'), ?)")
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, unable to prepare statement. %s", id, err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, data)
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, execute failed. %s", id, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Storage: PUT %s, the transaction commit failed. %s", id, err.Error())
	}

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

// ReadAll returns any item needs to be processed.
func (storage *SqliteStorage) ReadAll() ([][]byte, error) {
	var result [][]byte
	db, err := sql.Open("sqlite3", storage.dbPath)
	if err != nil {
		return nil, fmt.Errorf("Storage: READALL, open db failed. %s", err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT data FROM jobs ORDER BY addedAt")
	if err != nil {
		return nil, fmt.Errorf("Storage: READALL, query failed. %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("Storage: READALL, failed to scan. %s", err.Error())
		}
		result = append(result, data)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Storage: READALL, reading data failed. %s", err.Error())
	}

	return result, nil
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
