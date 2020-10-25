// This package monitors a folder and reruns any go files with main
// It also monitors for any new images and displays them
package main

import (
	"fmt"
	"hash/crc64"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/scottkirkwood/gart"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var (
	fileCrc         map[string]uint64
	fileCrcMux      sync.Mutex
	goFileToCompile string
	driverCount     int32
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Failed to create watcher: %v\n", err)
	}
	defer watcher.Close()

	fileCrc = make(map[string]uint64, 0)
	done := make(chan bool)
	newImgChan := make(chan string)

	folder := ""
	goFileToCompile, err = findMainGo(folder)
	if err != nil {
		fmt.Printf("Couldn't find go file with main in folder %q\n", folder)
		return
	}

	go watchForEvents(watcher, newImgChan)
	go maybeStartDriver(newImgChan)

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

func watchForEvents(watcher *fsnotify.Watcher, newImgChan chan string) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				go compileOne(event.Name)
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				go newFile(event.Name, newImgChan)
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

func newFile(fname string, newImgChan chan string) {
	fileChanged(fname) // changed is new, right?
	if !strings.HasSuffix(fname, ".png") {
		return
	}
	maybeStartDriver(newImgChan)
	newImgChan <- fname
}

var onlyDigitsRx = regexp.MustCompile(`\d+`)

func fileChanged(fname string) bool {
	if onlyDigitsRx.MatchString(fname) {
		// Ignore temp files by vim which have only digits
		return false
	}

	fileCrcMux.Lock()
	defer fileCrcMux.Unlock()

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

func maybeStartDriver(newImgChan chan string) {
	if !atomic.CompareAndSwapInt32(&driverCount, 0, 1) {
		fmt.Printf("driver already started\n")
		return
	}
	fmt.Printf("starting driver\n")
	driver.Main(func(s screen.Screen) {
		defer atomic.StoreInt32(&driverCount, 0)

		winSize := image.Point{1024, 768}

		w, err := s.NewWindow(&screen.NewWindowOptions{
			Width:  winSize.X,
			Height: winSize.Y,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		defer w.Release()

		// Watch the newImg chan and then them as events
		done := make(chan bool)
		go func(newImgChan chan string, done chan bool) {
			for {
				select {
				case img := <-newImgChan:
					w.Send(img)
				case <-done:
					return
				}
			}
		}(newImgChan, done)
		defer func() {
			done <- true
			close(done)
		}()

		b, err := s.NewBuffer(winSize)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer b.Release()

		w.Fill(b.Bounds(), color.White, draw.Src)
		w.Publish()
		var imgs []image.Image
		var sz size.Event
		var i int // index of image to display

		for {
			e := w.NextEvent()
			repaint := false
			switch e := e.(type) {
			case string:
				fmt.Printf("New image %s\n", e)
				_, imgs = gart.DecodeImages([]string{e})
				if len(imgs) > 0 {
					r := imgs[i].Bounds()
					sz.HeightPx = r.Dy()
					sz.WidthPx = r.Dx()
					repaint = true
					b, err = s.NewBuffer(sz.Size())
					if err != nil {
						fmt.Println(err)
						return
					}
					w.Publish()
				}
			case key.Event:
				switch e.Code {
				case key.CodeEscape, key.CodeQ:
					return
				case key.CodeR:
					if e.Direction == key.DirRelease {
						fmt.Printf("Refresh\n")
						if len(imgs) > 0 {
							// resize to current image
							r := imgs[i].Bounds()
							sz.HeightPx = r.Dy()
							sz.WidthPx = r.Dx()
							repaint = true
							b, err = s.NewBuffer(sz.Size())
							if err != nil {
								fmt.Println(err)
								return
							}
							w.Publish()
						}
					}
				}

			case paint.Event:
				if len(imgs) > 0 {
					img := imgs[i]
					draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)
					dp := gart.VpCenter(img, sz.WidthPx, sz.HeightPx)
					zero := image.Point{}
					if dp != zero {
						w.Fill(sz.Bounds(), color.Black, draw.Src)
					}
					w.Upload(dp, b, b.Bounds())
				}

			case size.Event:
				sz = e

			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case error:
				fmt.Printf("Screen error: %v\n", e)
				return

			case mouse.Event:
			default:
			}
			if repaint {
				w.Send(paint.Event{})
			}
		}
	})
}
