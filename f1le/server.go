package f1le

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	assets     string
	config     *Config
	store      *sessions.CookieStore
	uploadLock sync.Mutex
)

func Serve(port string, c *Config) error {
	if _, err := strconv.Atoi(port); err != nil {
		return errors.New("Invalid port number: " + port)
	}

	// Get configuration
	_, sourcePath, _, _ := runtime.Caller(0)
	assets = filepath.Join(filepath.Dir(filepath.Dir(sourcePath)), "assets")
	config = c
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
		securecookie.GenerateRandomKey(16))

	http.HandleFunc("/login", HandleLogin)
	http.HandleFunc("/upload", HandleUpload)
	http.HandleFunc("/", HandleRoot)
	return http.ListenAndServe(":"+port, nil)
}

type serverCtx struct {
	assets string
	config *Config
	store  *sessions.CookieStore
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		password := r.PostFormValue("password")
		if config.CheckPassword(password) {
			// Awesome, they logged in
			session, _ := store.Get(r, "sessid")
			session.Values["authenticated"] = true
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}
	
	log.Print("Serving login page")

	templatePath := filepath.Join(assets, "login.html")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, templatePath)
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Print("Static file: ", r.URL.Path)
		// Serve file
		path := strings.Replace(r.URL.Path, "..", "", -1)
		http.ServeFile(w, r, filepath.Join(assets, path))
		return
	}

	session, _ := store.Get(r, "sessid")
	if val, ok := session.Values["authenticated"].(bool); !ok || !val {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	
	log.Print("Serving homepage.")

	templatePath := filepath.Join(assets, "home.mustache")
	files := []string{"hey", "there", "bro", "yo", "bo", "ho", "yoyo"}
	template := map[string]interface{}{"files": files}
	body := mustache.RenderFile(templatePath, template)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Print("Invalid upload: not multipart")
		w.Write([]byte("{\"error\": \"Not multipart.\"}"))
		return
	}
	
	safeName := RandomString()
	part, err := reader.NextPart()
	if err != nil {
		log.Print("Invalid upload: missing part")
		w.Write([]byte("{\"error\": \"Missing part.\"}"))
		return
	}
	origName := part.FileName()
	localPath := filepath.Join(config.RootPath, safeName)
	output, err := os.Create(localPath)
	if err != nil {
		log.Print("Invalid upload: missing part")
		w.Write([]byte("{\"error\": \"Missing part.\"}"))
		return
	}
	io.Copy(output, part)
	part.Close()
	output.Close()
	uploadLock.Lock()
	config.Files[safeName] = origName
	uploadLock.Unlock()
	log.Print("Upload successful.")
	w.Write([]byte("{}"))
	return
}

func RandomString() string {
	randomNumber := strconv.Itoa(rand.Int()) + strconv.Itoa(rand.Int())
	hash := sha256.Sum256([]byte(randomNumber))
	return hex.EncodeToString(hash[:])
}
