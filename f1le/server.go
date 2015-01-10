package f1le

import (
	"errors"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	assets string
	config *Config
	store  *sessions.CookieStore
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

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/", RootHandler)
	return http.ListenAndServe(":"+port, nil)
}

type serverCtx struct {
	assets string
	config *Config
	store  *sessions.CookieStore
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	templatePath := filepath.Join(assets, "login.html")
	http.ServeFile(w, r, templatePath)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
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

	templatePath := filepath.Join(assets, "home.mustache")
	body := mustache.RenderFile(templatePath, map[string]interface{}{})
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}
