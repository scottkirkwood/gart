package gart

import (
	"fmt"
	"image"
	"os"
	"time"
)

// DecodeImages takes a list of image files and decodes them into image.Image
// types. Note that the number of images returned may not be the number of
// image files passed in. Namely, an image file is skipped if it cannot be
// read or deocoded into an image type that Go understands.
func DecodeImages(imageFiles []string) ([]string, []image.Image) {
	// A temporary type used to transport decoded images over channels.
	type tmpImage struct {
		img  image.Image
		name string
	}

	// Decoded all images specified in parallel.
	imgChans := make([]chan tmpImage, len(imageFiles))
	for i, fName := range imageFiles {
		imgChans[i] = make(chan tmpImage, 0)
		go func(i int, fName string) {
			file, err := os.Open(fName)
			if err != nil {
				fmt.Println(err)
				close(imgChans[i])
				return
			}

			start := time.Now()
			img, kind, err := image.Decode(file)
			if err != nil {
				fmt.Printf("Could not decode '%s' into a supported image "+
					"format: %s\n", fName, err)
				close(imgChans[i])
				return
			}
			fmt.Printf("Decoded '%s' into image type '%s' (%s).\n",
				fName, kind, time.Since(start))

			imgChans[i] <- tmpImage{
				img:  img,
				name: Basename(fName),
			}
		}(i, fName)
	}

	// Now collect all the decoded images into a slice of names and a slice
	// of images.
	names := make([]string, 0)
	imgs := make([]image.Image, 0)
	for _, imgChan := range imgChans {
		if tmpImg, ok := <-imgChan; ok {
			names = append(names, tmpImg.name)
			imgs = append(imgs, tmpImg.img)
		}
	}

	return names, imgs
}

// VpCenter inspects the canvas and image geometry, and determines where the
// origin of the image should be painted into the canvas.
// If the image is bigger than the canvas, this is always (0, 0).
// If the image is the same size, then it is also (0, 0).
// If a dimension of the image is smaller than the canvas, then:
// x = (canvas_width - image_width) / 2 and
// y = (canvas_height - image_height) / 2
func VpCenter(ximg image.Image, canWidth, canHeight int) image.Point {
	xmargin, ymargin := 0, 0
	if ximg.Bounds().Dx() < canWidth {
		xmargin = (canWidth - ximg.Bounds().Dx()) / 2
	}
	if ximg.Bounds().Dy() < canHeight {
		ymargin = (canHeight - ximg.Bounds().Dy()) / 2
	}
	return image.Point{xmargin, ymargin}
}
