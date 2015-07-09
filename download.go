package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"unicode"
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
		escapeNameForResult(file.Name))
	w.Header().Set("Content-Type", mimeTypeForName(file.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	io.Copy(w, f)
}

func HandleLast(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	DbLock.RLock()
	if len(Database.Files) == 0 {
		DbLock.RUnlock()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>No files</body></html>"))
		return
	}
	fileId := Database.Files[0].Id
	DbLock.RUnlock()

	http.Redirect(w, r, "/get/"+fileId, http.StatusTemporaryRedirect)
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

	// Set the headers and write the file
	w.Header().Set("Content-Disposition", "inline; filename="+
		escapeNameForResult(file.Name))
	w.Header().Set("Content-Type", mimeTypeForName(file.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	io.Copy(w, f)
}

func escapeNameForResult(filename string) string {
	res := ""
	for _, ch := range filename {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' {
			res += string(ch)
		} else {
			res += "-"
		}
	}
	return res
}

func mimeTypeForName(filename string) string {
	ext := path.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}
