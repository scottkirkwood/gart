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
	if !strings.HasSuffix(fname, ".png") {
		return
	}
	startDriver(fname)
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

func startDriver(fname string) {
	driver.Main(func(s screen.Screen) {
		// Decode all images (in parallel).
		_, imgs := gart.DecodeImages([]string{fname})

		// Return now if we don't have any images!
		if len(imgs) == 0 {
			fmt.Printf("No images specified could be shown.\n")
			return
		}

		// Auto-size the window with first image
		rect := imgs[0].Bounds()
		winSize := image.Point{rect.Dx(), rect.Dy()}
		if winSize.X > 1000 {
			winSize.X = 1000
		}
		if winSize.Y > 768 {
			winSize.Y = 768
		}

		w, err := s.NewWindow(&screen.NewWindowOptions{
			Width:  winSize.X,
			Height: winSize.Y,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		defer w.Release()

		b, err := s.NewBuffer(winSize)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer b.Release()

		w.Fill(b.Bounds(), color.White, draw.Src)
		w.Publish()

		var sz size.Event
		var i int // index of image to display
		for {
			e := w.NextEvent()
			switch e := e.(type) {
			case key.Event:
				repaint := false
				switch e.Code {
				case key.CodeEscape, key.CodeQ:
					return
				case key.CodeRightArrow:
					if e.Direction == key.DirPress {
						if i == len(imgs)-1 {
							i = -1
						}
						i++
						repaint = true
						b.Release()
						b, err = s.NewBuffer(sz.Size())
						if err != nil {
							fmt.Println(err)
							return
						}
					}

				case key.CodeLeftArrow:
					if e.Direction == key.DirPress {
						if i == 0 {
							i = len(imgs)
						}
						i--
						repaint = true
						b, err = s.NewBuffer(sz.Size())
						if err != nil {
							fmt.Println(err)
							return
						}
					}

				case key.CodeR:
					if e.Direction == key.DirPress {
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
				if repaint {
					w.Send(paint.Event{})
				}

			case paint.Event:
				img := imgs[i]
				draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)
				dp := gart.VpCenter(img, sz.WidthPx, sz.HeightPx)
				zero := image.Point{}
				if dp != zero {
					w.Fill(sz.Bounds(), color.Black, draw.Src)
				}
				w.Upload(dp, b, b.Bounds())

			case size.Event:
				sz = e

			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case error:
				fmt.Printf("Screen error: %v\n", e)
				return

			default:
			case mouse.Event:
			}
		}
	})
}
