package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
	securecookie.GenerateRandomKey(16))
var assetsPath string
var configuration *Config

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ", os.Args[0], " <port> <root path>")
	}
	
	// Setup configuration
	var err error
	_, sourcePath, _, _ := runtime.Caller(0)
	assetsPath = filepath.Join(filepath.Dir(sourcePath), "assets")
	configuration, err = LoadConfig(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	
	// Setup server
	if _, err := strconv.Atoi(os.Args[1]); err != nil {
		log.Fatal("Invalid port number: ", os.Args[1])
	}
	http.HandleFunc("/login", HandleLogin)
	http.HandleFunc("/", HandleHome)
	if err := http.ListenAndServe(":"+os.Args[1], nil); err != nil {
		log.Fatal("Error listening: ", err)
	}
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if (r.URL.Path != "/") {
		log.Print("Static file: ", r.URL.Path)
		// Serve file
		path := strings.Replace(r.URL.Path, "..", "", -1)
		http.ServeFile(w, r, filepath.Join(assetsPath, path))
		return
	}
	
	session, _ := store.Get(r, "sessid")
	if val, ok := session.Values["authenticated"].(bool); !ok || !val {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	
	templatePath := filepath.Join(assetsPath, "home.mustache")
	body := mustache.RenderFile(templatePath, map[string]interface{}{})
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		password := r.PostFormValue("password")
		hash := sha256.Sum256([]byte(password))
		hex := strings.ToLower(hex.EncodeToString(hash[:]))
		if hex == configuration.Hash {
			// Awesome, they logged in
			session, _ := store.Get(r, "sessid")
		    session.Values["authenticated"] = true
		    session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// TODO: use mustache to indicate they got the password wrong
	}
	
	templatePath := filepath.Join(assetsPath, "login.html")
	http.ServeFile(w, r, templatePath);
}

type Config struct {
	RootPath string `json:"-"`
	Hash     string
}

func LoadConfig(rootPath string) (*Config, error) {
	cfgPath := filepath.Join(rootPath, "config.json")
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		// Create a default configuration with password "password".
		cfg := &Config{rootPath,
			"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"}
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	res := &Config{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Config) Save() error {
	if data, err := json.Marshal(c); err != nil {
		return err
	} else {
		configPath := filepath.Join(c.RootPath, "config.json");
		return ioutil.WriteFile(configPath, data, os.FileMode(0700))
	}
}
