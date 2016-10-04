package main

import (
	"testing"
	"os"
)

const TestDBName string = "testdata/test.db"

func TestNewStorage(t *testing.T) {
	os.Remove(TestDBName)

	storage, err := NewStorage(TestDBName)
	if err != nil {
		t.Errorf("Unable to create new storage: %s", err.Error())
	}
	data := `{"url":"http://google.com"}`
	err = storage.Put("id1", []byte(data))
	if err != nil {
		t.Errorf("Unable to put data to new storage: %s", err.Error())
	}

	savedData, err := storage.Get("id1")
	if err != nil {
		t.Errorf("Unable to put data to new storage: %s", err.Error())
	}

	savedDataStr := string(savedData)

	if savedDataStr != "" && savedDataStr != data {
		t.Errorf("Expected %s is not equal to given %s", data, string(savedData))
	}

	err = storage.Remove("id1")
	if err != nil {
		t.Errorf("Unable to remove data from the storage: %s", err.Error())
	}

	savedData, err = storage.Get("id1")
	if err != nil {
		t.Errorf("Unable to get data from storage: %s", err.Error())
	}

	if savedData != nil {
		t.Errorf("Record was not removed from the storage")
	}
}
