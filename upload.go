package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
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

	// Perform the upload itself. Each file part is stored as a separate f1le.
	fileIds, err := UploadParts(reader)
	if err != nil {
		log.Print("Upload failed: ", err)
		if errors.Is(err, errMissingUpload) {
			writeUploadJSON(w, map[string]interface{}{"error": "Missing part."})
		} else {
			writeUploadJSON(w, map[string]interface{}{"error": "Upload failed."})
		}
		return
	}

	// Keep id for compatibility with existing single-file clients, while ids
	// reports every file stored by this request.
	writeUploadJSON(w, map[string]interface{}{
		"id":  fileIds[0],
		"ids": fileIds,
	})
}

func UploadStream(original string, r io.Reader) (string, error) {
	file, err := storeUpload(original, r)
	if err != nil {
		return "", err
	}
	if err := commitUploads([]File{file}); err != nil {
		os.Remove(filepath.Join(RootPath, file.Id))
		return "", err
	}
	return file.Id, nil
}

var errMissingUpload = errors.New("missing upload part")

// UploadParts stores every file in a multipart request. The files are added to
// the database as a single batch, in the same order as the multipart parts.
func UploadParts(reader *multipart.Reader) ([]string, error) {
	var files []File
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			removeUploads(files)
			return nil, err
		}

		// Ignore regular form values in case callers include other metadata.
		if part.FileName() == "" {
			part.Close()
			continue
		}
		file, err := storeUpload(part.FileName(), part)
		part.Close()
		if err != nil {
			removeUploads(files)
			return nil, err
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, errMissingUpload
	}

	if err := commitUploads(files); err != nil {
		removeUploads(files)
		return nil, err
	}
	ids := make([]string, len(files))
	for i, file := range files {
		ids[i] = file.Id
	}
	return ids, nil
}

func storeUpload(original string, r io.Reader) (File, error) {
	key := securecookie.GenerateRandomKey(16)
	fileId := strings.ToLower(hex.EncodeToString(key))
	localPath := filepath.Join(RootPath, fileId)
	output, err := os.Create(localPath)
	if err != nil {
		return File{}, err
	}

	size, err := io.Copy(output, r)
	closeErr := output.Close()
	if err != nil {
		os.Remove(localPath)
		return File{}, err
	}
	if closeErr != nil {
		os.Remove(localPath)
		return File{}, closeErr
	}
	return File{original, fileId, time.Now().UTC().Unix(), size}, nil
}

func commitUploads(files []File) error {
	DbLock.Lock()
	defer DbLock.Unlock()
	oldFiles := Database.Files
	newFiles := make([]File, 0, len(files)+len(oldFiles))
	newFiles = append(newFiles, files...)
	newFiles = append(newFiles, oldFiles...)
	Database.Files = newFiles
	if err := SaveDb(); err != nil {
		Database.Files = oldFiles
		return err
	}
	return nil
}

func removeUploads(files []File) {
	for _, file := range files {
		os.Remove(filepath.Join(RootPath, file.Id))
	}
}

func writeUploadJSON(w http.ResponseWriter, value interface{}) {
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Print("Failed to write upload response: ", err)
	}
}
