// L-system draws a fractal plant using a simple l-system
package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/fogleman/gg"
)

const (
	width    = 1500
	height   = 1500
	maxDepth = 7
)

func main() {
	ctx := gg.NewContext(width, height)
	ctx.SetColor(color.Gray{245})
	ctx.Clear()
	ctx.SetColor(color.Black)
	ctx.SetLineWidth(1)

	f := generateFractal(ctx, maxDepth)
	f.drawFractal()

	fname := "/tmp/output.png"
	fmt.Printf("Saved to %s\n", fname)
	ctx.SavePNG(fname)
}

type fractal struct {
	ctx        *gg.Context
	sequence   string
	maxDepth   int
	stack      []state
	minX, minY float64
	maxX, maxY float64
}

func generateFractal(ctx *gg.Context, depth int) *fractal {
	start := "x"
	for d := 0; d < depth; d++ {
		nextGeneration := make([]string, len(start))
		for i, c := range start {
			nextGeneration[i] = nextRule(c)
		}
		start = strings.Join(nextGeneration, "")
	}
	return &fractal{ctx: ctx, sequence: start, maxDepth: depth}
}

func nextRule(c rune) string {
	switch c {
	case 'f':
		return "ff"
	case 'x':
		return "x[-x]f+[[x]-x]-f[-fx]+x"
	default:
		return string(c)
	}
}

type state struct {
	x     float64
	y     float64
	angle float64
}

func (f *fractal) drawFractal() {
	f.internalDrawFractal(f.getLimits)
	fmt.Printf("Limits %f, %f\n", f.maxX, f.maxY)
	f.internalDrawFractal(f.drawTo)
	f.ctx.Stroke()
}

func (f *fractal) getLimits(s state) {
	f.minX = math.Min(s.x, f.minX)
	f.minY = math.Min(s.y, f.minY)
	f.maxX = math.Max(s.x, f.maxX)
	f.maxY = math.Max(s.y, f.maxY)
}

func (f *fractal) internalDrawFractal(drawTo func(state)) {
	s := state{}
	for _, c := range f.sequence {
		switch c {
		case 'f': // forward
			s.x += 5 * math.Cos(s.angle)
			s.y += 5 * math.Sin(s.angle)
			drawTo(s)
		case '-': // right
			s.angle += gg.Radians(8.75)
		case '+': // left
			s.angle -= gg.Radians(2.7)
		case '[': // push
			f.stack = append(f.stack, s)
		case ']': // pop
			s = f.stack[len(f.stack)-1]
			f.stack = f.stack[:len(f.stack)-1]
			f.moveTo(s)
		}
	}
}

func boundTo(min, max, toMin, toMax, x float64) float64 {
	return toMin + (x-min)*(toMax-toMin)/(max-min)
}

func (f *fractal) normX(x float64) float64 {
	return boundTo(f.minX, f.maxY, 0, width, x)
}

func (f *fractal) normY(y float64) float64 {
	return boundTo(f.minY, f.maxY, 0, height, y)
}

func (f *fractal) drawTo(s state) {
	f.ctx.SetLineWidth(2.0 / float64(f.maxDepth+1))
	x, y := f.normX(s.x), f.normY(s.y)
	f.ctx.LineTo(x, y)
	f.ctx.Stroke()
	f.ctx.MoveTo(x, y)
}

func (f *fractal) moveTo(s state) {
	f.ctx.ClearPath()
	f.ctx.MoveTo(f.normX(s.x), f.normY(s.y))
}
