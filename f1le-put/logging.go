package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var PrintLock sync.Mutex
var HasLogged bool

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
	Reader   io.Reader
	HasTotal bool
	Total    int64
	Current  int64
}

func (l *LogReader) Read(out []byte) (n int, err error) {
	PrintLock.Lock()
	defer PrintLock.Unlock()
	n, err = l.Reader.Read(out)
	l.Current += int64(n)
	if !l.HasTotal {
		fmt.Fprintf(os.Stderr, "\rread %d bytes", l.Current)
	} else {
		fmt.Fprintf(os.Stderr, "\rread %.2f%% (%d out of %d)",
			100*float64(l.Current)/float64(l.Total), l.Current, l.Total)
	}
	HasLogged = true
	return
}
