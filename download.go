package main

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
	"unicode"
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

	w.Header().Set("Content-Disposition", disposition+"; filename="+
		escapeNameForResult(file.Name))
	http.ServeContent(w, r, file.Name, time.Now(), f)

	return true
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
