package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Always reply with an empty JSON object.
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	// Get the input ID
	id := r.URL.Path[8:]
	if !ValidateId(id) {
		log.Print("Invalid delete request: ", id)
		return
	}

	log.Print("Deleting: ", id)

	DbLock.Lock()
	defer DbLock.Unlock()

	// Remove the entry from the database
	for i, file := range Database.Files {
		if file.Id == id {
			for j := i; j < len(Database.Files)-1; j++ {
				Database.Files[j] = Database.Files[j+1]
			}
			Database.Files = Database.Files[0 : len(Database.Files)-1]
		}
	}
	SaveDb()

	// Remove the local file
	if err := os.Remove(filepath.Join(RootPath, id)); err != nil {
		log.Print("Failed to delete ", id, ": ", err)
		return
	}
}
