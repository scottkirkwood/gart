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
	"sort"

	"github.com/scottkirkwood/gart"
)

const (
	width            = 215.9 // mm
	height           = 279.4 // mm
	defaultLineWidth = 0.3
	startingCracks   = 10

	dimx = 1024 // pixels
	dimy = 768

	maxnum = 500
	maxPal = 512
)

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

	s := newSubstrate(dimx, dimy, maxnum, palette)
	if err := g.SafeWrite(ctx, "substrate-", ".png"); err != nil {
		fmt.Printf("Unable write image: %v\n", err)
		return
	}
}

type Substrate struct {
	cgrid              []int
	cracks             []*Crack
	goodcolor          color.Palette
	dimx, dimy, maxnum int
	sp                 *SandPainter
}

func newSubstrate(dimx, dimy, maxnum int, palette color.Palette) Substrate {
	return Substrate{
		cgrid:     make([]int, dimy*dimx),
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
	for y := 0; y < dimy; y++ {
		for x := 0; x < dimx; x++ {
			s.cgrid[y*dimx+x] = 10001
		}
	}
	// make random crack seeds
	for k := 0; k < 16; k++ {
		i := int(rand.Intn(dimx*dimy - 1))
		s.cgrid[i] = rand.Intn(360)
	}

	// make just three cracks
	for k := 0; k < startingCracks; k++ {
		s.makeCrack()
	}
	//background(255);
}

func (s *Substrate) makeCrack() {
	// make a new crack instance
	if len(s.cracks) < cap(s.cracks) {
		s.cracks = append(s.cracks, newCrack())
	}
}

type Crack struct {
	x, y float64
	t    float64 // direction of travel in degrees
}

func newCrack() *Crack {
	return &Crack{}
}

func (c *Crack) regionColor(s *Substrate) {
	// start checking one step away
	rx := c.x
	ry := c.y
	openspace := true

	// find extents of open space
	for openspace {
		// move perpendicular to crack
		sin, cos := math.Sincos(c.t * math.Pi / 180)
		rx += 0.81 * sin
		ry -= 0.81 * cos
		cx := int(rx)
		cy := int(ry)
		if cx >= 0 && cx < dimx && cy >= 0 && cy < dimy {
			// safe to check
			if s.cgrid[cy*s.dimx+cx] > 10000 {
				// space is open
			} else {
				openspace = false
			}
		} else {
			openspace = false
		}
	}
	// draw sand painter
	s.sp.render(rx, ry, c.x, c.y)
}

func (c *Crack) findStart(s *Substrate) {
	// pick random point
	px := 0
	py := 0

	// shift until crack is found
	found := false
	for timeout := 0; timeout < 1000; timeout++ {
		px = rand.Intn(dimx)
		py = rand.Intn(dimy)
		if s.cgrid[py*dimx+px] < 10000 {
			found = true
			break
		}
	}

	if found {
		// start crack
		a := float64(s.cgrid[py*dimx+px])
		if rand.Intn(100) < 50 {
			a -= 90 + rand.Float64()*4.1 - 2
		} else {
			a += 90 + rand.Float64()*4.1 - 2
		}
		c.startCrack(float64(px), float64(py), a)
	}
}

func (c *Crack) startCrack(X, Y, T float64) {
	c.x = X
	c.y = Y
	c.t = T //%360;
	sin, cos := math.Sincos(c.t * math.Pi / 180)
	c.x += 0.61 * cos
	c.y += 0.61 * sin
}

func (c *Crack) move(s *Substrate) {
	// continue cracking
	sin, cos := math.Sincos(c.t * math.Pi / 180)
	c.x += 0.42 * cos
	c.y += 0.42 * sin

	// bound check
	z := 0.33
	cx := int(c.x + rand.Float64()*2*z - z) // add fuzz
	cy := int(c.y + rand.Float64()*2*z - z)

	// draw sand painter
	c.regionColor(s)

	// draw black crack
	gart.Stroke(0, 85)
	gart.Point(c.x+randRange(-z, z), c.y+randRange(-z, z))

	if (cx >= 0) && (cx < s.dimx) && (cy >= 0) && (cy < s.dimy) {
		// safe to check
		if s.cgrid[cy*s.dimx+cx] > 10000 || math.Abs(float64(s.cgrid[cy*s.dimx+cx]))-c.t < 5 {
			// continue cracking
			s.cgrid[cy*s.dimx+cx] = int(c.t)
		} else if math.Abs(float64(s.cgrid[cy*s.dimx+cx])-c.t) > 3 {
			// crack encountered (not self), stop cracking
			c.findStart(s)
			s.makeCrack()
		}
	} else {
		// out of bounds, stop cracking
		c.findStart(s)
		s.makeCrack()
	}
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

func (s *SandPainter) render(x, y, ox, oy float64) {
	// modulate gain
	s.g += randRange(-0.050, 0.050)
	maxg := 1.0
	if s.g < 0 {
		s.g = 0
	}
	if s.g > maxg {
		s.g = maxg
	}

	// calculate grains by distance
	//int grains = int(sqrt((ox-x)*(ox-x)+(oy-y)*(oy-y)));
	grains := 64.0

	// lay down grains of sand (transparent pixels)
	w := s.g / (grains - 1)
	for i := 0.0; i < grains; i++ {
		a := 0.1 - i/(grains*10.0)
		gart.Stroke(s.RGBA())
		gart.Point(ox+(x-ox)*math.Sin(math.Sin(i*w)), oy+(y-oy)*math.Sin(math.Sin(i*w)))
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
	sort.Slice(toSort, func(i, j int) bool { return toSort[i].count > toSort[j].count })

	pal := make(color.Palette, 0, len(toSort))
	for _, colCount := range toSort {
		pal = append(pal, colCount.col)
	}
	return pal[:maxPal], nil
}

func randRange(low, high float64) float64 {
	return rand.Float64()*(high-low) + low
}
