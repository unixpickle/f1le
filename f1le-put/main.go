package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
)

func main() {
	var uploadFileName string

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: f1le-put [flags] [path]\n\n"+
			"Set F1LE_ROOT and F1LE_PASS environment variables.\n"+
			"F1LE_ROOT is a URL, like http://localhost:1337.\n\n"+
			"Available flags:")
		flag.PrintDefaults()
	}
	flag.StringVar(&uploadFileName, "name", "", "upload file name")
	flag.Parse()

	var sourceFile *os.File

	switch len(flag.Args()) {
	case 0:
		if uploadFileName == "" {
			uploadFileName = "stdin"
		}
		sourceFile = os.Stdin
		fmt.Fprintln(os.Stderr, "reading from standard input.")
	case 1:
		path := flag.Args()[0]
		var err error
		sourceFile, err = os.Open(path)
		if err != nil {
			essentials.Die(err)
		}
		if uploadFileName == "" {
			uploadFileName = filepath.Base(path)
		}
		defer sourceFile.Close()
	default:
		dieUsage()
	}

	baseURL, password := readEnv()

	jar, err := cookiejar.New(nil)
	if err != nil {
		dieError(err)
	}
	client := &http.Client{Jar: jar}

	authenticate(client, *baseURL, password)
	resp := postFile(client, *baseURL, sourceFile, uploadFileName)
	defer resp.Body.Close()
	printResponse(*baseURL, resp)
}

func readEnv() (baseURL *url.URL, password string) {
	baseStr := os.Getenv("F1LE_ROOT")
	password = os.Getenv("F1LE_PASS")
	if baseStr == "" || password == "" {
		dieUsage()
	}
	var err error
	baseURL, err = url.Parse(baseStr)
	if err != nil {
		dieError("invalid F1LE_ROOT:", baseStr)
	}
	return
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

func postFile(c *http.Client, u url.URL, f *os.File, name string) *http.Response {
	var fileSize int64
	if f != os.Stdin {
		var err error
		fileSize, err = f.Seek(0, io.SeekEnd)
		if err != nil {
			dieError(err)
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			dieError(err)
		}
	}

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		dieError(err)
	}

	mpWriter := multipart.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		defer mpWriter.Close()
		fileWriter, err := mpWriter.CreateFormFile("file-input", name)
		if err != nil {
			dieError("upload failed:", err)
		}
		lr := &LogReader{
			Reader:   f,
			HasTotal: f != os.Stdin,
			Total:    fileSize,
		}
		if _, err := io.Copy(fileWriter, lr); err != nil {
			dieError("upload failed:", err)
		}
	}()

	u.Path = "/upload"
	req, _ := http.NewRequest("POST", u.String(), pipeReader)
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())

	resp, err := c.Do(req)
	if err != nil {
		dieError("upload failed:", err)
	}
	return resp
}

func printResponse(u url.URL, resp *http.Response) {
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

func dieUsage() {
	flag.Usage()
	os.Exit(1)
}
