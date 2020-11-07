package gart

import (
	"image/color"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/pdf"
	"github.com/tdewolff/canvas/rasterizer"
	"github.com/tdewolff/canvas/svg"
)

// Context is my abstraction for Canvas (or gg)
type Context struct {
	c   *canvas.Canvas
	ctx *canvas.Context
}

func NewContext(width, height float64) *Context {
	ctx := &Context{
		c: canvas.New(width, height),
	}
	ctx.ctx = canvas.NewContext(ctx.c)
	return ctx
}

// WritePNG writes to a PNG file
func (ctx *Context) WritePNG(fname string) error {
	return ctx.c.WriteFile(fname, rasterizer.PNGWriter(3.2))
}

// WriteSVG writes to an SVG file
func (ctx *Context) WriteSVG(fname string) error {
	return ctx.c.WriteFile(fname, svg.Writer)
}

// WritePDF writes to a PDF file
func (ctx *Context) WritePDF(fname string) error {
	return ctx.c.WriteFile(fname, pdf.Writer)
}

func (ctx *Context) Push() {
	ctx.ctx.Push()
}

// Pop restores the last pushed draw state and uses that as the current draw state. If there are no
// states on the stack, this will do nothing.
func (ctx *Context) Pop() {
	ctx.ctx.Pop()
}

// Reset empties the canvas.
func (ctx *Context) Reset() {
	ctx.c.Reset()
}

func (ctx *Context) SetFillColor(col color.Color) {
	ctx.ctx.SetFillColor(col)
}

func (ctx *Context) SetStrokeColor(col color.Color) {
	ctx.ctx.SetStrokeColor(col)
}

func (ctx *Context) SetStrokeWidth(width float64) {
	ctx.ctx.SetStrokeWidth(width)
}

// MoveTo moves the path to x,y without connecting the path. It starts a new independent subpath.
// Multiple subpaths can be useful when negating parts of a previous path by overlapping it with a
// path in the opposite direction. The behaviour for overlapping paths depend on the FillRule.
func (ctx *Context) MoveTo(x, y float64) {
	ctx.ctx.MoveTo(x, y)
}

// Point draws a 1 pixel rectangle at point
func (ctx *Context) Point(x, y float64) {
	ctx.ctx.DrawPath(x, y, canvas.Rectangle(1, 1))
}

// LineTo adds a linear path to x,y.
func (ctx *Context) LineTo(x, y float64) {
	ctx.ctx.LineTo(x, y)
}

// QuadTo adds a quadratic Bézier path with control point cpx,cpy and end point x,y.
func (ctx *Context) QuadTo(cpx, cpy, x, y float64) {
	ctx.ctx.QuadTo(cpx, cpy, x, y)
}

// FillRect draws a rectable path
func (ctx *Context) FillRect(x, y, w, h float64) {
	ctx.ctx.DrawPath(x, y, canvas.Rectangle(w, h))
}

// Stroke strokes the current path and resets it.
func (ctx *Context) Stroke() {
	ctx.ctx.Stroke()
}

// Close closes the current path
func (ctx *Context) Close() {
	ctx.ctx.Close()
}
