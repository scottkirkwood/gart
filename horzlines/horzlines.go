//
// Inspired by github.com/bcongdon/generative-doodles/1-27-19.png
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/scottkirkwood/gart"
)

const (
	width            = 215.9 // mm
	height           = 279.4 // mm
	defaultLineWidth = 0.6
	cols             = 200
	deltaY           = 10  // pixels
	muteY            = 0.2 // std of 1/2 height variation
)

var seedFlag = flag.String("seed", "", "Hex value for the seed to use")

func main() {
	flag.Parse()
	g, err := gart.Init(*seedFlag)
	if err != nil {
		fmt.Printf("Unable to set the seed: %v\n", err)
	}

	ctx := gart.NewContext(width, height)
	ctx.SetFillColor(color.Gray{245})
	ctx.FillRect(0, 0, width, height)
	ctx.SetStrokeColor(color.Black)
	ctx.SetStrokeWidth(defaultLineWidth)

	draw(ctx)

	if err := g.SafeWrite(ctx, "horzlines-", ".png"); err != nil {
		fmt.Printf("Unable write image: %v\n", err)
		return
	}
}

func draw(ctx *gart.Context) {
	ypoints := make([]float64, cols)
	deltaX := float64(width / cols)

	rc := colorful.Hsl(30.0+rand.Float64()*50.0, 0.2+rand.Float64()*0.8, 0.3+rand.Float64()*0.7)
	hue, sat, light := rc.Hsl()
	ctx.SetStrokeColor(rc)

	maxDy := 0.0
	for y := 0.0; y < float64(height)+maxDy; y += float64(deltaY) {
		ctx.MoveTo(0, float64(y))
		for i := 0; i < cols; i++ {
			ypoints[i] += rand.NormFloat64() * deltaY * muteY
			maxDy = math.Max(maxDy, ypoints[i])

		}
		rc = colorful.Hsl(hue, gart.Lerp(sat, 1, y/(height+maxDy)), light)
		ctx.SetStrokeColor(rc)
		for i := 0; i < cols; i += 2 {
			if i >= cols-2 {
				break
			}
			x := float64(i)
			//ctx.LineTo(x*deltaX, y+ypoints[i])
			ctx.QuadTo(
				(x+0)*deltaX, y+ypoints[i+0],
				(x+1)*deltaX, y+ypoints[i+1])
			//ctx.CubicTo(
			//	(x+0)*deltaX, y+ypoints[i+0],
			//	(x+1)*deltaX, y+ypoints[i+1],
			//	(x+2)*deltaX, y+ypoints[i+2])
		}
		ctx.Stroke()
		ctx.Close()
	}
}
