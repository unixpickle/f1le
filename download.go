package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[5:]
	if !serveFile(w, r, id, "attachment") {
		http.NotFound(w, r)
	}
}

func HandleView(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[6:]
	if !serveFile(w, r, id, "inline") {
		http.NotFound(w, r)
	}
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

func serveFile(w http.ResponseWriter, r *http.Request, id, disposition string) bool {
	if !ValidateId(id) {
		log.Println("Invalid download ID:", id)
		return false
	}

	log.Println("Serving:", id)

	file, ok := findFileForId(id)
	if !ok {
		return false
	}

	f, err := os.Open(filepath.Join(RootPath, id))
	if err != nil {
		log.Println("Failed to open file for:", id)
		return false
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", dispositionHeader(disposition, file.Name))
	http.ServeContent(w, r, file.Name, time.Now(), f)

	return true
}

func dispositionHeader(disposition, filename string) string {
	return disposition + "; filename*=UTF-8''" + url.PathEscape(filename)
}

func findFileForId(id string) (file File, found bool) {
	DbLock.RLock()
	defer DbLock.RUnlock()
	for _, f := range Database.Files {
		if f.Id == id {
			found = true
			file = f
			return
		}
	}
	return
}
