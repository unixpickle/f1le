package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type LogWriter struct {
	Writer  io.Writer
	Total   int64
	Current int64
	Wrote   bool
}

func (l *LogWriter) Write(out []byte) (n int, err error) {
	n, err = l.Writer.Write(out)
	l.Current += int64(n)
	if l.Total == 0 {
		fmt.Fprintf(os.Stderr, "\rdownloaded %d bytes", l.Current)
	} else {
		fmt.Fprintf(os.Stderr, "\rdownloaded %.2f%% (%d out of %d)",
			100*float64(l.Current)/float64(l.Total), l.Current, l.Total)
	}
	l.Wrote = true
	return
}

func dieUsage() {
	flag.Usage()
	os.Exit(1)
}
