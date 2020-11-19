// Stolen from Substrate Watercolor, J Tarbell, June 2004
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	//"sort"

	"github.com/scottkirkwood/gart"
)

const (
	width            = 215.9 // mm
	height           = 279.4 // mm
	defaultLineWidth = 0.3
	startingCracks   = 10

	dimx = 1024 // pixels
	dimy = 768

	maxnum     = 500
	maxPal     = 512
	emptyAngle = -1
)

type degrees int

var (
	seedFlag = flag.String("seed", "", "Hex value for the seed to use")
)

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

	palette, err := takeColors("pollockShimmering.jpg")
	fmt.Printf("Num colors %d\n", len(palette))
	if err != nil {
		fmt.Printf("Unable get palette: %v\n", err)
		return
	}
	s := newSubstrate(ctx, dimx, dimy, maxnum, palette)
	s.begin()
	s.makeCrack()
	s.draw()

	if true {
		if err := g.SafeWrite(ctx, "substrate-", ".png"); err != nil {
			fmt.Printf("Unable write image: %v\n", err)
			return
		}
	} else {
		if err := g.SafeWrite(ctx, "substrate-", ".svg"); err != nil {
			fmt.Printf("Unable write image: %v\n", err)
			return
		}
	}
}

type Substrate struct {
	ctx   *gart.Context
	cgrid []degrees

	cracks             []*Crack
	goodcolor          color.Palette
	dimx, dimy, maxnum int
	sp                 *SandPainter
}

func newSubstrate(ctx *gart.Context, dimx, dimy, maxnum int, palette color.Palette) Substrate {
	return Substrate{
		ctx:       ctx,
		cgrid:     make([]degrees, dimy*dimx),
		cracks:    make([]*Crack, 0, maxnum),
		goodcolor: palette,
		dimx:      dimx,
		dimy:      dimy,
		maxnum:    maxnum,
		sp:        newSandPainter(palette),
	}
}

func (s *Substrate) begin() {
	// erase crack grid
	for y := 0; y < s.dimy; y++ {
		for x := 0; x < s.dimx; x++ {
			s.setAngle(x, y, emptyAngle)
		}
	}
	// make random crack seeds
	for k := 0; k < 6; k++ {
		x := rand.Intn(s.dimx)
		y := rand.Intn(s.dimy)
		s.setAngle(x, y, degrees(rand.Intn(360)))
	}

	// make just three cracks
	for k := 0; k < startingCracks; k++ {
		s.makeCrack()
	}
	//background(255);
}

func (s *Substrate) setAngle(x, y int, deg degrees) {
	s.cgrid[y*s.dimx+x] = deg
}

func (s *Substrate) getAngle(x, y int) degrees {
	return s.cgrid[y*s.dimx+x]
}

func (s *Substrate) inBounds(x, y int) bool {
	return x >= 0 && x < s.dimx && y >= 0 && y < s.dimy
}

func (s *Substrate) draw() {
	for i := 0; i < 3000; i++ {
		// crack all cracks
		for n := 0; n < len(s.cracks); n++ {
			s.cracks[n].move(s)
		}
	}
	//s.ctx.Close()
}

func (s *Substrate) makeCrack() {
	// make a new crack instance
	if len(s.cracks) < cap(s.cracks) {
		s.cracks = append(s.cracks, newCrack(s))
	}
}

type Crack struct {
	x, y float64
	t    float64 // direction of travel in degrees
}

func newCrack(s *Substrate) *Crack {
	// find placement along existing crack
	c := &Crack{}
	c.findStart(s)
	return c
}

func (c *Crack) regionColor(s *Substrate) {
	// start checking one step away
	rx := c.x
	ry := c.y

	// find extents of open space
	for {
		// move perpendicular to crack
		sin, cos := math.Sincos(gart.Radians(c.t))
		rx += 0.81 * sin
		ry -= 0.81 * cos
		cx := int(rx)
		cy := int(ry)
		if s.inBounds(cx, cy) {
			// safe to check
			if s.getAngle(cx, cy) == emptyAngle {
				// space is open
			} else {
				break
			}
		} else {
			break
		}
	}
	// draw sand painter
	s.sp.render(s.ctx, rx, ry, c.x, c.y)
}

func (c *Crack) findStart(s *Substrate) {
	px, py, found := s.findRandomPoint()
	if found {
		// start crack
		ang := float64(s.getAngle(px, py))
		randDeg := 90 + rand.Float64()*4.1 - 2
		if rand.Intn(100) < 50 {
			ang -= randDeg
		} else {
			ang += randDeg
		}
		c.startCrack(s, float64(px), float64(py), ang)
	}
}

func (s *Substrate) findRandomPoint() (x, y int, found bool) {
	for timeout := 0; timeout < 1000; timeout++ {
		px := rand.Intn(s.dimx)
		py := rand.Intn(s.dimy)
		if s.getAngle(px, py) != emptyAngle {
			return px, py, true
		}
	}
	return 0, 0, false
}

func (c *Crack) startCrack(s *Substrate, X, Y, T float64) {
	c.x = X
	c.y = Y
	c.t = math.Mod(T, 360)
	sin, cos := math.Sincos(gart.Radians(c.t))
	c.x += 0.61 * cos
	c.y += 0.61 * sin

	s.ctx.MoveTo(c.x, c.y)
}

func (c *Crack) move(s *Substrate) {
	// continue cracking
	sin, cos := math.Sincos(gart.Radians(c.t))
	c.x += 0.42 * cos
	c.y += 0.42 * sin

	// bound check
	z := 0.33
	cx := int(c.x + rand.Float64()*2*z - z) // add fuzz
	cy := int(c.y + rand.Float64()*2*z - z)

	// draw sand painter
	c.regionColor(s)

	// draw black crack
	s.ctx.SetStrokeColor(color.RGBA{0, 0, 0, 85})
	s.ctx.LineTo(c.x+randRange(-z, z), c.y+randRange(-z, z))

	if s.inBounds(cx, cy) {
		// safe to check
		if s.getAngle(cx, cy) >= emptyAngle ||
			math.Abs(float64(s.getAngle(cx, cy)))-c.t < 5 {
			// continue cracking
			s.setAngle(cx, cy, degrees(c.t))
		} else if math.Abs(float64(s.getAngle(cx, cy))-c.t) > 3 {
			// crack encountered (not self), stop cracking
			c.findStart(s)
			s.makeCrack()
		}
	} else {
		// out of bounds, stop cracking
		c.findStart(s)
		s.makeCrack()
	}
	s.ctx.Stroke()
}

type SandPainter struct {
	c color.Color // Color
	g float64     // Grain
}

func newSandPainter(palette color.Palette) *SandPainter {
	return &SandPainter{
		c: palette[rand.Intn(len(palette))],
		g: randRange(0.01, 0.1),
	}
}

func (sp *SandPainter) render(ctx *gart.Context, x, y, ox, oy float64) {
	// modulate grain
	sp.g += randRange(-0.050, 0.050)
	maxg := 1.0
	if sp.g < 0 {
		sp.g = 0
	}
	if sp.g > maxg {
		sp.g = maxg
	}

	// calculate grains by distance
	//int grains = int(sqrt((ox-x)*(ox-x)+(oy-y)*(oy-y)));
	grains := 64.0

	// lay down grains of sand (transparent pixels)
	w := sp.g / (grains - 1)
	ctx.MoveTo(ox, oy)
	for i := 0.0; i < grains; i++ {
		aa := 0.1 - i/(grains*10.0)
		rr, gg, bb, _ := sp.c.RGBA()
		ctx.SetStrokeColor(color.RGBA{
			R: uint8(rr),
			G: uint8(gg),
			B: uint8(bb),
			A: uint8(aa * 256)})
		siniw := math.Sin(math.Sin(i * w))
		if true {
			ctx.LineTo(
				ox+(x-ox)*siniw,
				oy+(y-oy)*siniw)
		}
	}
}

func takeColors(fname string) (color.Palette, error) {
	reader, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	bounds := m.Bounds()
	colorMap := make(map[color.Color]int, 512)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			col := m.At(x, y)
			colorMap[col]++
		}
	}
	type colCount struct {
		col   color.Color
		count int
	}
	toSort := make([]colCount, 0, len(colorMap))
	for key, val := range colorMap {
		toSort = append(toSort, colCount{key, val})
	}
	//sort.Slice(toSort, func(i, j int) bool { return toSort[i].count > toSort[j].count })

	pal := make(color.Palette, 0, len(toSort))
	for _, colCount := range toSort {
		pal = append(pal, colCount.col)
	}
	return pal[:maxPal], nil
}

func randRange(low, high float64) float64 {
	return rand.Float64()*(high-low) + low
}

func showPalette(ctx *gart.Context, palette color.Palette, w, h float64) {
	rows := 16
	cols := len(palette) / rows
	dx := w / float64(cols)
	dy := h / float64(rows)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			index := y*cols + x
			if index >= len(palette) {
				break
			}
			ctx.SetFillColor(palette[index])
			ctx.FillRect(float64(x)*dx, float64(y)*dy, dx, dy)
		}
	}
}
