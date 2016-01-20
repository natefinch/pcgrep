package main

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

var errlog = log.New(os.Stderr, "", 0)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}
	pat := os.Args[1]
	reg, err := regexp.Compile(pat)
	if err != nil {
		errlog.Println("Error compiling regex: %v", err)
		os.Exit(1)
	}
	files := os.Args[2:]

	var readers sync.WaitGroup
	for _, f := range files {
		readers.Add(1)
		go read(f, &readers, reg)
	}
	readers.Wait()
}

func usage() {
	log.Printf(`
usage:
%s <regex> <file> [<file>...]
`, filepath.Base(os.Args[0]))
}

func read(name string, wg *sync.WaitGroup, reg *regexp.Regexp) {
	defer wg.Done()
	f, err := os.Open(name)
	if err != nil {
		errlog.Println(err)
		return
	}
	defer f.Close()
	var r io.Reader
	ext := filepath.Ext(name)
	switch ext {
	case ".gz":
		r, err = gzip.NewReader(f)
		if err != nil {
			log.Println(err)
			return
		}
	case ".bz":
		r = bzip2.NewReader(f)
	default:
		errlog.Println("Unknown extension:", ext)
		return
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		if reg.Match(b) {
			log.Printf("%s: %s", name, b)
		}
	}
	if err := scanner.Err(); err != nil {
		errlog.Printf("Error while reading %s: %s", name, err)
	}
}
