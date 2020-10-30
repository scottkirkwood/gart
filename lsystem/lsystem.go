// L-system draws a fractal 'plant' using a simple l-system like logo
// Inspired by github.com/bcongdon/generative-doodles/blob/master/2-27-19
// Seed is currently unused (fix!)
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"

	"github.com/scottkirkwood/gart"
)

const (
	width            = 215.9 // mm
	height           = 279.4 // mm
	defaultLineWidth = 0.3
	maxDepth         = 7
)

var (
	seedFlag = flag.String("seed", "", "Hex value for the seed to use")
	lsystems = []lsystem{
		{
			name:       "Tree Like",
			startAngle: 90,
			angles:     []float64{8.78, 2.7},
			depth:      7,
			dist:       6, // pixels
			axiom:      "x",
			rules: map[rune]string{
				'f': "ff",
				'x': "x[-x]f+[[x]-x]-f[-fx]+x",
			},
			lineWidth: func(curDepth, maxDepth int) float64 {
				return float64(maxDepth-2) / math.Pow(float64(curDepth+1), 1.6)
			},
		}, {
			name:       "Hilbert",
			startAngle: 0,
			angles:     []float64{90},
			depth:      6,
			dist:       2, // pixels
			axiom:      "l",
			rules: map[rune]string{
				'l': "→+rf-lfl-fr+",
				'r': "-lf+rfr+fl-",
			},
			lineWidth: func(curDepth, maxDepth int) float64 { return defaultLineWidth },
		}, {
			name:       "Sierpinski Gasket",
			startAngle: 180,
			angles:     []float64{45},
			depth:      8,
			dist:       1,
			axiom:      "x",
			rules: map[rune]string{
				'x': "yf+xf+y",
				'y': "xf-yf-x",
			},
			lineWidth: func(curDepth, maxDepth int) float64 { return defaultLineWidth },
		}, {
			name:       "Gosper curve",
			startAngle: 90,
			angles:     []float64{60},
			depth:      5,
			dist:       1,
			axiom:      "x",
			rules: map[rune]string{
				'x': "x+yf++yf-fx--fxfx-yf+",
				'y': "-fx+yfyf++yf+fx--fx-y",
			},
			lineWidth: func(curDepth, maxDepth int) float64 { return defaultLineWidth },
		}, {
			name:       "Plant 1",
			startAngle: 90,
			angles:     []float64{20},
			depth:      5,
			dist:       1,
			axiom:      "x",
			rules: map[rune]string{
				'f': "ff",
				'x': "f+[-f-xf-x][+ff][--xf[+x]][++f-x]",
			},
			lineWidth: func(curDepth, maxDepth int) float64 { return defaultLineWidth },
		}, {
			name:       "Plant 2",
			startAngle: 90,
			angles:     []float64{15},
			depth:      5,
			dist:       1,
			axiom:      "X",
			rules: map[rune]string{
				'F': "FX[FX[+XF]]",
				'X': "FF[+XZ++X-F[+ZX]][-X++F-X]",
				'Z': "[+F-X-F][++ZX]",
			},
			lineWidth: func(curDepth, maxDepth int) float64 { return float64(maxDepth-curDepth)/5.9 + defaultLineWidth/2 },
		}, {
			name:       "Rectangles",
			startAngle: 90,
			angles:     []float64{90},
			depth:      4,
			dist:       1,
			axiom:      "F-F-F-F",
			rules: map[rune]string{
				'F': "FF-F+F-F-FF",
			},
		},
	}
)

type lsystem struct {
	name       string
	startAngle float64   // degrees 90 is up
	angles     []float64 // degrees
	depth      int
	dist       float64
	axiom      string
	rules      map[rune]string
	lineWidth  func(curDepth, maxDepth int) float64
}

type fractal struct {
	ctx                   *gart.Context
	sequence              string
	lsys                  lsystem
	width, height         int // pixels
	stack                 []turtle
	angleLeft, angleRight float64 // radians
	minX, minY            float64
	maxX, maxY            float64
}

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

	lsystem := lsystems[rand.Intn(len(lsystems))]
	lsystem = lsystems[len(lsystems)-2]

	f := initFractal(ctx, lsystem,
		int(math.Floor(width)),
		int(math.Floor(height)))
	f.generate()
	f.draw()

	if err := g.SafeWrite(ctx, "lsystem-", ".png"); err != nil {
		fmt.Printf("Unable write image: %v\n", err)
		return
	}
}

func initFractal(ctx *gart.Context, lsys lsystem, width, height int) *fractal {
	angleLeft := lsys.angles[0]
	angleRight := angleLeft
	if len(lsys.angles) == 2 {
		angleRight = lsys.angles[1]
	}

	if lsys.lineWidth == nil {
		lsys.lineWidth = func(curDepth, maxDepth int) float64 { return defaultLineWidth }
	}
	f := &fractal{
		ctx:       ctx,
		lsys:      lsys,
		angleLeft: gart.Radians(angleLeft), angleRight: gart.Radians(angleRight),
		width: width, height: height,
	}
	return f
}

type turtle struct {
	x     float64
	y     float64
	angle float64
}

func (f *fractal) generate() string {
	var ok bool
	start := f.lsys.axiom
	for d := 0; d < f.lsys.depth; d++ {
		nextGeneration := make([]string, len(start))
		for i, c := range start {
			nextGeneration[i], ok = f.lsys.rules[c]
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
	// Turtle starts facing up
	s := turtle{angle: gart.Radians(f.lsys.startAngle)}
	for _, c := range f.sequence {
		switch c {
		case 'f', 'F', 'G', 'g': // move forward
			dy, dx := math.Sincos(s.angle)
			s.y += f.lsys.dist * dy
			s.x += f.lsys.dist * dx
			drawTo(s, len(f.stack))
		case '-': // turn left
			s.angle += f.angleLeft
		case '+': // turn right
			s.angle -= f.angleRight
		case '[': // push
			f.stack = append(f.stack, s)
		case ']': // pop!
			s = f.stack[len(f.stack)-1]
			f.stack = f.stack[:len(f.stack)-1]
			f.moveTo(s, len(f.stack))
		}
	}
	f.ctx.Close()
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
	f.ctx.SetStrokeWidth(f.lsys.lineWidth(depth, f.lsys.depth))
	x, y := f.normX(s.x), f.normY(s.y)
	f.ctx.LineTo(x, y)
	f.ctx.Stroke()
	f.ctx.MoveTo(x, y)
}

func (f *fractal) moveTo(s turtle, depth int) {
	f.ctx.Close()
	f.ctx.MoveTo(f.normX(s.x), f.normY(s.y))
}
