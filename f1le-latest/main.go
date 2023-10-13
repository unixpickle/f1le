package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/f1le/cliutil"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: f1le-latest [path]")
		fmt.Fprintln(os.Stderr, "")
		cliutil.PrintEnvUsage()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Available flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	var outFile io.Writer

	switch len(flag.Args()) {
	case 0:
		outFile = os.Stdout
	case 1:
		path := flag.Args()[0]
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		essentials.Must(err)
		outFile = f
		defer f.Close()
	default:
		dieUsage()
	}

	client, baseURL, err := cliutil.AuthEnvClient()
	essentials.Must(err)
	baseURL.Path = "/last"
	resp, err := client.Get(baseURL.String())

	// Log fraction of written size
	size, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	lw := &LogWriter{Writer: outFile, Total: size}
	outFile = lw

	essentials.Must(err)
	defer resp.Body.Close()
	_, err = io.Copy(outFile, resp.Body)
	essentials.Must(err)

	if lw.Wrote {
		fmt.Fprintln(os.Stderr)
	}
}
