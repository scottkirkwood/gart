// This package monitors a folder and reruns any go files with main
// It also monitors for any new images and displays them
package main

import (
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var (
	fileCrc         map[string]uint64
	goFileToCompile string
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Failed to create watcher: %v\n", err)
	}
	defer watcher.Close()

	fileCrc = make(map[string]uint64, 0)
	done := make(chan bool)

	folder := ""
	goFileToCompile, err = findMainGo(folder)
	if err != nil {
		fmt.Printf("Couldn't find go file with main in folder %q\n", folder)
		return
	}

	go watchForEvents(watcher)

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(folder); err != nil {
		fmt.Printf("Problem add folder watcher: %v\n", err)
	}
	if folder == "" {
		folder, _ = os.Getwd()
	}
	fmt.Printf("Monitoring folder %q\n", folder)

	<-done
}

func watchForEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				go compileOne(event.Name)
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				go newFile(event.Name)
			} else if event.Op&fsnotify.Rename == fsnotify.Write {
				//fmt.Println("renamed file:", event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("ERROR", err)
		}
	}
}

func compileOne(fname string) {
	if !fileChanged(fname) {
		if strings.HasSuffix(fname, ".go") {
			fmt.Printf("File %s unchanged\n", fname)
		}
		return
	}
	if !strings.HasSuffix(fname, ".go") {
		return
	}
	fmt.Printf("Running %s\n", goFileToCompile)
	cmdName := "/usr/bin/go"
	cmdArgs := []string{"run", goFileToCompile}
	if cmdOut, err := exec.Command(cmdName, cmdArgs...).CombinedOutput(); err != nil {
		fmt.Printf("Err: %v: %s\n", err, cmdOut)
	} else {
		fmt.Printf("Compiled!: %s\n", cmdOut)
	}
}

func newFile(fname string) {
	// Add new file to the list?
	fileChanged(fname)
}

var onlyDigitsRx = regexp.MustCompile(`\d+`)

func fileChanged(fname string) bool {
	if onlyDigitsRx.MatchString(fname) {
		// Ignore temp files by vim which have only digits
		return false
	}
	checksum := fileCrc[fname]
	newChecksum := fileChecksum(fname)
	if newChecksum == checksum {
		return false
	}
	fileCrc[fname] = newChecksum
	return true
}

func fileChecksum(fname string) uint64 {
	h := crc64.New(crc64.MakeTable(crc64.ECMA))
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Readfile error %q: %v\n", fname, err)
		return 0
	}
	h.Write(bytes)
	return h.Sum64()
}

// findMainGo searches for a go file that has "package main"
func findMainGo(folder string) (string, error) {
	if folder == "" {
		folder = "."
	}
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		found, err := hasMain(f.Name())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if found {
			return f.Name(), nil
		}
	}
	return "", nil
}

var packageMainRx = regexp.MustCompile(`\s?package main\s?`)

func hasMain(fname string) (bool, error) {
	if !strings.HasSuffix(fname, ".go") {
		// Not a go file
		return false, nil
	}
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Readfile error %q: %v\n", fname, err)
		return false, err
	}
	return packageMainRx.Match(bytes), nil
}
