package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Db struct {
	Hash  string `json:"hash"`
	Files []File `json:"files"`
}

type File struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	Uploaded int64  `json:"uploaded"`
	Size     int64  `json:"size"`
}

var (
	RootPath   string
	AssetsPath string
	Store      *sessions.CookieStore
	Database   *Db
	DbLock     sync.RWMutex
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) != 3 {
		log.Fatal("Usage: f1le <port> <root path>")
	}

	// Setup global variables
	_, sourcePath, _, _ := runtime.Caller(0)
	AssetsPath = filepath.Join(filepath.Dir(sourcePath), "assets")
	RootPath = os.Args[2]
	Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
		securecookie.GenerateRandomKey(16))

	// Load database
	if err := LoadDb(); err != nil {
		log.Fatal(err)
	}

	// Setup the server
	http.HandleFunc("/delete/", HandleDelete)
	http.HandleFunc("/get/", HandleDownload)
	http.HandleFunc("/files", HandleFiles)
	http.HandleFunc("/login", HandleLogin)
	http.HandleFunc("/upload", HandleUpload)
	http.HandleFunc("/", HandleRoot)
	log.Print("Attempting to listen on http://localhost:" + os.Args[1])
	if err := http.ListenAndServe(":"+os.Args[1], nil); err != nil {
		log.Fatal(err)
	}
}

func FileTemplate(f File) map[string]string {
	// Compute a human-readable file size
	var sizeString string
	if f.Size < 1024 {
		sizeString = strconv.FormatInt(f.Size, 10) + " bytes"
	} else if f.Size < 1048576 {
		sizeString = strconv.FormatInt(f.Size/1024, 10) + " KiB"
	} else if f.Size < 1073741824 {
		sizeString = strconv.FormatInt(f.Size/1048576, 10) + " MiB"
	} else {
		sizeString = strconv.FormatInt(f.Size/1073741824, 10) + " GiB"
	}
	
	// Create a human-readable date
	t := time.Unix(f.Uploaded, 0)
	dateString := strconv.Itoa(int(t.Month())) + "/" + strconv.Itoa(t.Day()) +
		"/" + strconv.Itoa(t.Year())
	
	// Find the appropriate image
	icon := "unknown"
	icons := map[string]string{
		"png":  "image",
		"jpg":  "image",
		"jpeg": "image",
		"gif":  "image",
	}
	lowerName := strings.ToLower(f.Name)
	for ext, img := range icons {
		if strings.HasSuffix(lowerName, "." + ext) {
			icon = img
			break
		}
	}
	
	return map[string]string{
		"name":     f.Name,
		"id":       f.Id,
		"uploaded": dateString,
		"size":     sizeString,
		"icon":     "images/icons/" + icon + ".png",
	}
}

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
				Database.Files[j] = Database.Files[j + 1]
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
	w.Header().Set("Content-Disposition", "attachment; filename=" +
		url.QueryEscape(file.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	io.Copy(w, f)
}

func HandleFiles(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// We don't want their browser to cache the file list.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	DbLock.RLock()
	fileMaps := make([]map[string]string, len(Database.Files))
	for i, file := range Database.Files {
		fileMaps[i] = FileTemplate(file)
	}
	DbLock.RUnlock()

	template := map[string]interface{}{"files": fileMaps}
	templatePath := filepath.Join(AssetsPath, "files.mustache")
	body := mustache.RenderFile(templatePath, template)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{"failed": false}
	if r.Method == "POST" {
		password := r.PostFormValue("password")
		DbLock.RLock()
		authed := (HashPassword(password) == Database.Hash)
		DbLock.RUnlock()
		if authed {
			s, _ := Store.Get(r, "sessid")
			s.Values["authenticated"] = true
			s.Save(r, w)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		log.Print("Failed login attempt.")
		template["failed"] = true
	}

	log.Print("Serving login page.")

	templatePath := filepath.Join(AssetsPath, "login.mustache")
	body := mustache.RenderFile(templatePath, template)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	// This handler will be called for all static files.
	if r.URL.Path != "/" {
		cleanPath := strings.Replace(r.URL.Path, "..", "", -1)
		log.Print("Static file: ", cleanPath)
		http.ServeFile(w, r, filepath.Join(AssetsPath, cleanPath))
		return
	}

	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	http.ServeFile(w, r, filepath.Join(AssetsPath, "index.html"))
}

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

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

func IsAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	s, _ := Store.Get(r, "sessid")
	val, ok := s.Values["authenticated"].(bool)
	return ok && val
}

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

func ValidateId(id string) bool {
	for _, ch := range id {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
			return false
		}
	}
	return true
}
