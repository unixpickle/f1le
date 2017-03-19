package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	if len(os.Args) != 2 {
		dieUsage()
	}
	baseURL := os.Getenv("F1LE_ROOT")
	password := os.Getenv("F1LE_PASS")
	if baseURL == "" || password == "" {
		dieUsage()
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		dieError(err)
	}
	client := &http.Client{Jar: jar}

	u, err := url.Parse(baseURL)
	if err != nil {
		dieError("invalid F1LE_ROOT:", baseURL)
	}
	u.Path = "/upload"

	authenticate(client, *u, password)

	f, err := os.Open(os.Args[1])
	if err != nil {
		dieError(err)
	}
	defer f.Close()

	fileSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		dieError(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		dieError(err)
	}

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		dieError(err)
	}

	mpWriter := multipart.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		defer mpWriter.Close()
		name := filepath.Base(os.Args[1])
		fileWriter, err := mpWriter.CreateFormFile("file-input", name)
		if err != nil {
			dieError("upload failed:", err)
		}
		lr := &LogReader{r: f, total: fileSize}
		if _, err := io.Copy(fileWriter, lr); err != nil {
			dieError("upload failed:", err)
		}
	}()

	req, _ := http.NewRequest("POST", u.String(), pipeReader)
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		dieError("upload failed:", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		dieError("upload failed:", err)
	}

	var respObj struct {
		ID    string `json:"id"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		dieError("unexpected response")
	}
	if respObj.Error != "" {
		dieError("remote error:", respObj.Error)
	}

	u.Path = "/get/" + url.PathEscape(respObj.ID)
	fmt.Fprintln(os.Stderr, "")
	fmt.Println(u.String())
}

func authenticate(c *http.Client, u url.URL, password string) {
	u.Path = "/login"
	vals := url.Values{}
	vals.Set("password", password)
	resp, err := c.PostForm(u.String(), vals)
	if err != nil {
		dieError("authentication failure:", err)
	}
	resp.Body.Close()
	if resp.Request.URL.Path == "/login" {
		dieError("login failed")
	}
}

var PrintLock sync.Mutex
var HasLogged bool

func dieUsage() {
	dieError("Usage: f1le-put <path>\n\n" +
		"Set F1LE_ROOT and F1LE_PASS env variables.\n" +
		"F1LE_ROOT is a URL, like http://localhost:1337.")
}

func dieError(args ...interface{}) {
	PrintLock.Lock()
	defer PrintLock.Unlock()
	if HasLogged {
		fmt.Fprintln(os.Stderr, "")
	}
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

type LogReader struct {
	r     io.Reader
	total int64
	cur   int64
}

func (l *LogReader) Read(out []byte) (n int, err error) {
	PrintLock.Lock()
	defer PrintLock.Unlock()
	n, err = l.r.Read(out)
	l.cur += int64(n)
	fmt.Fprintf(os.Stderr, "\rread %.2f%% (%d out of %d)",
		100*float64(l.cur)/float64(l.total), l.cur, l.total)
	HasLogged = true
	return
}
