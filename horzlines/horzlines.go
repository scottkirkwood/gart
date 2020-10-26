//
// Inspired by github.com/bcongdon/generative-doodles/1-27-19.png
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/fogleman/gg"
	"github.com/scottkirkwood/gart"
)

const (
	width  = 1024 // pixels
	height = 768  // pixels
	cols   = 100
	deltaY = 10  // pixels
	muteY  = 0.2 // std of 1/2 height variation
)

var seedFlag = flag.String("seed", "", "Hex value for the seed to use")

func main() {
	flag.Parse()
	g, err := gart.Init(*seedFlag)
	if err != nil {
		fmt.Printf("Unable to set the seed: %v\n", err)
	}

	ctx := gg.NewContext(width, height)
	ctx.SetColor(color.Gray{245})
	ctx.Clear()
	ctx.SetColor(color.Black)
	ctx.SetLineWidth(2)

	draw(ctx)

	if err := g.SafeWrite(ctx, "horzlines-", ".png"); err != nil {
		fmt.Printf("Unable write image: %v\n", err)
		return
	}
}

func draw(ctx *gg.Context) {
	ypoints := make([]float64, cols)
	deltaX := float64(width / cols)

	maxDy := 0.0
	for y := 0.0; y < float64(height)+maxDy; y += float64(deltaY) {
		ctx.MoveTo(0, float64(y))
		for i := 0; i < cols; i++ {
			ypoints[i] += rand.NormFloat64() * deltaY * muteY
			maxDy = math.Max(maxDy, ypoints[i])

		}
		for i := 0; i < cols; i += 2 {
			if i >= cols-2 {
				break
			}
			x := float64(i)
			//ctx.LineTo(x*deltaX, y+ypoints[i])
			ctx.QuadraticTo(
				(x+0)*deltaX, y+ypoints[i+0],
				(x+1)*deltaX, y+ypoints[i+1])
			//ctx.CubicTo(
			//	(x+0)*deltaX, y+ypoints[i+0],
			//	(x+1)*deltaX, y+ypoints[i+1],
			//	(x+2)*deltaX, y+ypoints[i+2])
		}
		ctx.Stroke()
		ctx.ClearPath()
	}
}
