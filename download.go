package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Get the input ID
	id := r.URL.Path[5:]
	for !ValidateId(id) {
		log.Print("Invalid download ID: ", id)
		http.NotFound(w, r)
		return
	}

	log.Print("Downloading: ", id)

	DbLock.RLock()
	defer DbLock.RUnlock()

	// Find the file in the database
	var file File
	found := false
	for _, f := range Database.Files {
		if f.Id == id {
			found = true
			file = f
		}
	}
	if !found {
		log.Print("Not found for download: ", id)
		http.NotFound(w, r)
		return
	}

	// Open the file
	f, err := os.Open(filepath.Join(RootPath, id))
	if err != nil {
		log.Print("Failed to open file for: ", id)
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	// Set the headers and write the file
	w.Header().Set("Content-Disposition", "attachment; filename="+
		url.QueryEscape(file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	io.Copy(w, f)
}

func HandleView(w http.ResponseWriter, r *http.Request) {
	// Get the input ID
	id := r.URL.Path[6:]
	for !ValidateId(id) {
		log.Print("Invalid download ID: ", id)
		http.NotFound(w, r)
		return
	}

	log.Print("Viewing: ", id)

	DbLock.RLock()
	defer DbLock.RUnlock()

	// Find the file in the database
	var file File
	found := false
	for _, f := range Database.Files {
		if f.Id == id {
			found = true
			file = f
		}
	}
	if !found {
		log.Print("Not found for view: ", id)
		http.NotFound(w, r)
		return
	}

	// Open the file
	f, err := os.Open(filepath.Join(RootPath, id))
	if err != nil {
		log.Print("Failed to open file for: ", id)
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	
	// Compute the MIME type
	ext := path.Ext(file.Name)
	mimeType := mime.TypeByExtension(ext)
	log.Print(mimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Set the headers and write the file
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	io.Copy(w, f)
}
