package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	Database *Db
	DbLock   sync.RWMutex
)

func LoadDb() error {
	dbPath := filepath.Join(RootPath, "data.json")
	data, err := ioutil.ReadFile(dbPath)
	if err == nil {
		Database = &Db{}
		if err := json.Unmarshal(data, Database); err != nil {
			return err
		}
		return nil
	}

	// Create a new database my prompting them for a password.
	fmt.Print("Please enter a new password: ")
	var password string
	fmt.Scanln(&password)
	Database = &Db{HashPassword(password), []File{}}
	if err := SaveDb(); err != nil {
		return err
	}
	return nil
}

func SaveDb() error {
	if data, err := json.Marshal(Database); err != nil {
		return err
	} else {
		dbPath := filepath.Join(RootPath, "data.json")
		return ioutil.WriteFile(dbPath, data, os.FileMode(0700))
	}
}
