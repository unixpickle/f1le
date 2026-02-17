package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/f1le/cliutil"
	"github.com/unixpickle/seektar"
)

func main() {
	var uploadFileName string

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: f1le-put [flags] [path]")
		fmt.Fprintln(os.Stderr, "")
		cliutil.PrintEnvUsage()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Available flags:")
		flag.PrintDefaults()
	}
	flag.StringVar(&uploadFileName, "name", "", "upload file name")
	flag.Parse()

	var sourceFile io.ReadSeekCloser

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
		var basename string
		sourceFile, basename, err = openFileOrTar(path)
		if err != nil {
			essentials.Die(err)
		}
		if uploadFileName == "" {
			uploadFileName = basename
		}
		defer sourceFile.Close()
	default:
		dieUsage()
	}

	client, baseURL, err := cliutil.AuthEnvClient()
	if err != nil {
		dieError(err)
	}
	resp := postFile(client, *baseURL, sourceFile, uploadFileName)
	defer resp.Body.Close()
	printResponse(*baseURL, resp)
}

func openFileOrTar(path string) (io.ReadSeekCloser, string, error) {
	// Make the path absolute to get a basename for a relative path
	// like "." or "..".
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, "", err
	}
	basename := filepath.Base(path)

	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}

	if !info.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			return nil, "", err
		}
		return f, basename, nil
	} else {
		fmt.Fprintln(os.Stderr, "uploading directory as a TAR archive.")
		agg, err := seektar.Tar(path, basename)
		if err != nil {
			return nil, "", err
		}
		f, err := agg.Open()
		if err != nil {
			return nil, "", err
		}
		return f, basename + ".tar", nil
	}
}

func postFile(c *http.Client, u url.URL, f io.ReadSeeker, name string) *http.Response {
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
			Reader: f,
			Total:  fileSize,
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
	body, err := io.ReadAll(resp.Body)
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
	printForTerminalOrPipe(u.String())
}

func printForTerminalOrPipe(data string) {
	fi, err := os.Stdout.Stat()
	if err != nil {
		dieError("failed to stat stdout:", err)
	}
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println(data)
	} else {
		// For convenience when piping into a command like `pbcopy`,
		// we do not write a newline so that this path can be pasted
		// directly into part of a command.
		fmt.Print(data)
	}
}

func dieUsage() {
	flag.Usage()
	os.Exit(1)
}
