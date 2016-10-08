package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

const TestDBName string = "testdata/test.db"

func TestNewStorage(t *testing.T) {
	os.Remove(TestDBName)

	storage, err := NewStorage(TestDBName)
	if err != nil {
		t.Errorf("[TestNewStorage] Unable to create new storage: %s", err.Error())
	}
	data := `{"url":"http://google.com"}`
	err = storage.Put("id1", []byte(data))
	if err != nil {
		t.Errorf("[TestNewStorage] Unable to put data to new storage: %s", err.Error())
	}

	savedData, err := storage.Get("id1")
	if err != nil {
		t.Errorf("[TestNewStorage] Unable to put data to new storage: %s", err.Error())
	}

	savedDataStr := string(savedData)

	if savedDataStr != "" && savedDataStr != data {
		t.Errorf("[TestNewStorage] Expected %s is not equal to given %s", data, string(savedData))
	}

	err = storage.Remove("id1")
	if err != nil {
		t.Errorf("[TestNewStorage] Unable to remove data from the storage: %s", err.Error())
	}

	savedData, err = storage.Get("id1")
	if err != nil {
		t.Errorf("[TestNewStorage] Unable to get data from storage: %s", err.Error())
	}

	if savedData != nil {
		t.Errorf("[TestNewStorage] Record was not removed from the storage")
	}
}

func TestReadAll(t *testing.T) {
	os.Remove(TestDBName)

	storage, err := NewStorage(TestDBName)
	if err != nil {
		t.Errorf("[ReadAll] Unable to create new storage: %s", err.Error())
	}
	data := [...][]byte{
		[]byte(`{"url":"http://google.com"}`),
		[]byte(`{"url":"http://yahoo.com"}`),
		[]byte(`{"url":"http://bing.com"}`),
	}
	dataLen := len(data)

	for i := 0; i < dataLen; i++ {
		err = storage.Put(fmt.Sprintf("id%s", i), data[i])
		if err != nil {
			t.Errorf("[ReadAll] Unable to put data to the storage: %s", err.Error())
		}
	}

	results, err := storage.ReadAll()
	if err != nil {
		t.Errorf("[ReadAll] Unable to read all: %s", err.Error())
	}
	if len(results) != dataLen {
		t.Errorf("[ReadAll] results length Expected=%s;Actual=%s;", dataLen, len(results))
	} else {
		for i := 0; i < dataLen; i++ {
			found := false
			for j := 0; j < dataLen; j++ {
				if bytes.Compare(results[i], data[j]) == 0 {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[ReadAll] %s not found", string(data[i]))
			}
		}
	}
}
