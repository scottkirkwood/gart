// L-system draws a fractal 'plant' using a simple l-system like logo
// Inspired by github.com/bcongdon/generative-doodles/blob/master/2-27-19
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"github.com/scottkirkwood/gart"
)

const (
	width      = 1200                // pixels
	height     = width * (728 / 383) // pixels
	maxDepth   = 7
	dist       = 6    // pixels
	angleLeft  = 8.78 // degrees
	angleRight = 2.7  // ditto
)

var (
	seedFlag = flag.String("seed", "", "Hex value for the seed to use")
	rules    = []string{
		"f→ ff",
		"x→ x[-x]f+[[x]-x]-f[-fx]+x",
	}
)

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
	ctx.SetLineWidth(1)

	f := initFractal(ctx, dist, angleLeft, angleRight, rules, width, height, maxDepth)
	f.generate()
	f.draw()

	if err := g.SafeWrite(ctx, "lsystem-", ".png"); err != nil {
		fmt.Printf("Unable write image: %v\n", err)
		return
	}
}

type fractal struct {
	ctx                   *gg.Context
	sequence              string
	maxDepth              int
	dist                  float64 // how far to move forward, pixels
	angleLeft, angleRight float64 // radians
	width, height         int     // pixels
	stack                 []turtle
	rules                 map[rune]string
	minX, minY            float64
	maxX, maxY            float64
}

func initFractal(ctx *gg.Context, dist, angleLeft, angleRight float64, rules []string, width, height, depth int) *fractal {
	f := &fractal{
		ctx:       ctx,
		dist:      dist,
		rules:     make(map[rune]string, len(rules)),
		angleLeft: gg.Radians(angleLeft), angleRight: gg.Radians(angleRight),
		width: width, height: height,
		maxDepth: depth,
	}
	for _, rule := range rules {
		parts := strings.Split(rule, "→ ")
		char := []rune(parts[0])[0]
		f.rules[char] = parts[1]
	}
	return f
}

type turtle struct {
	x     float64
	y     float64
	angle float64
}

func (f *fractal) generate() string {
	start := "x"
	var ok bool
	for d := 0; d < f.maxDepth; d++ {
		nextGeneration := make([]string, len(start))
		for i, c := range start {
			nextGeneration[i], ok = f.rules[c]
			if !ok {
				nextGeneration[i] = string(c)
			}
		}
		start = strings.Join(nextGeneration, "")
	}
	f.sequence = start
	return f.sequence
}

func (f *fractal) draw() {
	f.internalGenerate(f.getLimits)
	fmt.Printf("Limits %f, %f\n", f.maxX, f.maxY)
	f.internalGenerate(f.drawTo)
	f.ctx.Stroke()
}

func (f *fractal) getLimits(s turtle, _ int) {
	f.minX = math.Min(s.x, f.minX)
	f.minY = math.Min(s.y, f.minY)
	f.maxX = math.Max(s.x, f.maxX)
	f.maxY = math.Max(s.y, f.maxY)
}

func (f *fractal) internalGenerate(drawTo func(turtle, int)) {
	s := turtle{angle: gg.Radians(90)}
	for _, c := range f.sequence {
		switch c {
		case 'f': // move forward
			s.x += f.dist * math.Cos(s.angle)
			s.y += f.dist * math.Sin(s.angle)
			drawTo(s, len(f.stack))
		case '-': // turn left
			s.angle += f.angleLeft
		case '+': // turn right
			s.angle -= f.angleRight
		case '[': // push
			f.stack = append(f.stack, s)
		case ']': // pop
			s = f.stack[len(f.stack)-1]
			f.stack = f.stack[:len(f.stack)-1]
			f.moveTo(s, len(f.stack))
		}
	}
}

func boundTo(foundMin, foundMax, canvasMin, canvasMax, x float64) float64 {
	return canvasMin + (x-foundMin)*(canvasMax-canvasMin)/(foundMax-foundMin)
}

func (f *fractal) normX(x float64) float64 {
	return boundTo(f.minX, f.maxX, 0, float64(f.width), x)
}

func (f *fractal) normY(y float64) float64 {
	return boundTo(f.minY, f.maxY, 0, float64(f.height), y)
}

func (f *fractal) drawTo(s turtle, depth int) {
	f.ctx.SetLineWidth(float64(f.maxDepth-2) / math.Pow(float64(depth+1), 1.6))
	x, y := f.normX(s.x), f.normY(s.y)
	f.ctx.LineTo(x, y)
	f.ctx.Stroke()
	f.ctx.MoveTo(x, y)
}

func (f *fractal) moveTo(s turtle, depth int) {
	f.ctx.ClearPath()
	f.ctx.MoveTo(f.normX(s.x), f.normY(s.y))
}
