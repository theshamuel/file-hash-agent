package fileagent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileAgent struct {
	Path     string
	Interval time.Duration
	Delay    time.Duration

	wg sync.WaitGroup
}

func (s *FileAgent) Run() error {
	s.wg.Add(1)
	msg := make(chan string)
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
	go getListFiles(ctx, s.Path, &s.wg, msg)
	go func(c chan string) {
		s.wg.Wait()
		close(c)
	}(msg)
	var files []string
	for r := range msg {
		files = append(files, r)
	}
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		h.Write([]byte(f[strings.Index(f, "#fha#")+5:]))
	}
	log.Printf("[DEBUG]: hash str = %s", hex.EncodeToString(h.Sum(nil)))
	return nil
}

func getListFiles(ctx context.Context, dir string, wg *sync.WaitGroup, res chan string) {
	defer wg.Done()
	file, err := os.Open(dir)
	if err != nil {
		fmt.Println("error opening directory")
	}
	defer file.Close()

	files, err := file.Readdir(0)
	if err != nil {
		fmt.Println("error reading directory")
	}
	for _, f := range files {
		if f.IsDir() {
			wg.Add(1)
			go getListFiles(ctx, dir+"/"+f.Name(), wg, res)
		} else {
			dir := dir + "/"
			log.Printf("[DEBUG]: [goroutine id = %s] filepath = %s%s", getGoroutineId(), dir, f.Name())
			res <- strings.TrimPrefix(dir, "./_example/") + f.Name()
			wg.Add(1)
			go checksum(dir+f.Name(), wg, res)
		}

	}
}

func getGoroutineId() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	gid := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return gid
}

func checksum(file string, wg *sync.WaitGroup, res chan<- string) {
	defer wg.Done()
	f, err := os.Open(file)
	if err != nil {
		log.Panic("[ERROR] can not open file")
	}

	defer f.Close()

	copyBuf := make([]byte, 1024*1024)

	h := sha256.New()
	if _, err := io.CopyBuffer(h, f, copyBuf); err != nil {
		log.Printf("[ERROR] %v\n", err)
	}

	res <- file + "#fha#" + hex.EncodeToString(h.Sum(nil))
}
