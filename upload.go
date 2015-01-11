package main

import (
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
)

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// This is a JSON+AJAX API.
	w.Header().Set("Content-Type", "application/json")

	// Get the multipart reader.
	reader, err := r.MultipartReader()
	if err != nil {
		log.Print("Invalid upload: not multipart")
		w.Write([]byte("{\"error\": \"Not multipart.\"}"))
		return
	}

	// There should only be one part, and that part should contain the file.
	part, err := reader.NextPart()
	if err != nil {
		log.Print("Invalid upload: missing part")
		w.Write([]byte("{\"error\": \"Missing part.\"}"))
		return
	}

	// Perform the upload itself.
	if fileId, err := UploadStream(part.FileName(), part); err != nil {
		log.Print("Upload failed: ", err)
		w.Write([]byte("{\"error\": \"Upload failed.\"}"))
	} else {
		w.Write([]byte("{\"id\": \"" + fileId + "\"}"))
	}
}

func UploadStream(original string, r io.Reader) (string, error) {
	key := securecookie.GenerateRandomKey(16)
	fileId := strings.ToLower(hex.EncodeToString(key))
	localPath := filepath.Join(RootPath, fileId)
	output, err := os.Create(localPath)
	if err != nil {
		return "", err
	}

	size, err := io.Copy(output, r)
	output.Close()
	if err != nil {
		os.Remove(localPath)
		return "", err
	}

	// Create a new File and insert it at the front of the list.
	DbLock.Lock()
	defer DbLock.Unlock()
	file := File{original, fileId, time.Now().UTC().Unix(), size}
	Database.Files = append([]File{file}, Database.Files...)
	if err := SaveDb(); err != nil {
		Database.Files = Database.Files[1:]
		os.Remove(localPath)
		return "", err
	} else {
		return fileId, nil
	}
}
