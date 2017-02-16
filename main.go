package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/hoisie/mustache"
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

	// Load database
	if err := LoadDb(); err != nil {
		log.Fatal(err)
	}

	// Setup the server
	handlers := map[string]http.HandlerFunc{
		"/delete/": HandleDelete,
		"/get/":    HandleDownload,
		"/files":   HandleFiles,
		"/login":   HandleLogin,
		"/logout":  HandleLogout,
		"/upload":  HandleUpload,
		"/last":    HandleLast,
		"/view/":   HandleView,
		"/":        HandleRoot,
	}
	for path, f := range handlers {
		http.Handle(path, context.ClearHandler(f))
	}
	log.Print("Attempting to listen on http://localhost:" + os.Args[1])
	if err := http.ListenAndServe(":"+os.Args[1], nil); err != nil {
		log.Fatal(err)
	}
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

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	s, _ := Store.Get(r, "sessid")
	s.Values["authenticated"] = false
	s.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
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

func ValidateId(id string) bool {
	for _, ch := range id {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
			return false
		}
	}
	return true
}
