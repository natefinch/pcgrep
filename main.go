package main

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 3 {
		usage()
		return
	}
	pat := os.Args[1]
	reg, err := regexp.Compile(pat)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error compiling regex: %v", err)
		os.Exit(1)
	}
	files := os.Args[2:]
	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go grep(f, reg, &wg)
	}
	wg.Wait()
}

func usage() {
	log.Printf(`
usage:
%s <regex> <file> [<file>...]
`, filepath.Base(os.Args[0]))
}

func grep(name string, reg *regexp.Regexp, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.Open(name)
	if err != nil {
		log.Println(err)
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
		fmt.Fprintln(os.Stderr, "*** Unknown extension:", ext)
		return
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		if reg.Match(b) {
			log.Printf("%s: %s", name, string(b))
		}
	}
}
