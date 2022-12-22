package main

import (
	"image/color"
	"math"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"

	"gioui.org/font/gofont"
)

func main() {
	var program App
	go func() {
		w := app.NewWindow(app.Title("Clock"))
		program.loop(w) // returns when window is closed
		os.Exit(0)
	}()
	app.Main()
}

type App struct{}

func (app *App) loop(w *app.Window) {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops

	for e := range w.Events() {
		if e, ok := e.(system.FrameEvent); ok {
			gtx := layout.NewContext(&ops, e)
			app.layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

var (
	secondColor  = color.NRGBA{A: 0xff, R: 0xD8, G: 0xff}
	minuteColor  = color.NRGBA{A: 0xff, R: 0x80, G: 0xff}
	hourColor    = color.NRGBA{A: 0xff, G: 0xFF}
	dayColor     = color.NRGBA{A: 0xff, B: 0xC0, G: 0x80}
	yearColor    = color.NRGBA{A: 0xff, R: 0x20, B: 0xE0, G: 0x60}
	centuryColor = color.NRGBA{A: 0xff, B: 0xFF, R: 0x40}
)

func (app *App) layout(gtx layout.Context, th *material.Theme) {
	now := time.Now()

	// Convert the current time into the fraction of the arc for each segment.
	// For the segments that are part of the current day, we first figure out how
	// much time has passed since the start of the current day and then figure out
	// what fraction of each second/min/hour/day we have.
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayElapsed := now.Sub(dayStart)
	secondFraction := (dayElapsed % time.Second).Seconds()
	minuteFraction := (dayElapsed % time.Minute).Seconds() / 60
	hourFraction := (dayElapsed % time.Hour).Seconds() / 3600
	dayFraction := dayElapsed.Seconds() / (24 * 60 * 60)

	// For year and century we have to separately figure out the appropriate
	// starting time and then determine the completed fraction.
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	yearElapsed := now.Sub(yearStart)
	yearFraction := yearElapsed.Seconds() / (365 * 24 * 60 * 60)

	centuryStart := time.Date(now.Year()-now.Year()%100, time.January, 1, 0, 0, 0, 0, now.Location())
	centuryElapsed := now.Sub(centuryStart)
	centuryFraction := centuryElapsed.Seconds() / (100 * 365 * 24 * 60 * 60)

	// Center of the window
	cx := float32(gtx.Constraints.Max.X) / 2
	cy := float32(gtx.Constraints.Max.Y) / 2
	// Radius of each segment -- the min of x/y split into 6 parts. We'll actually
	// use the first 5 parts and only a bit of the 6th for the second angle.
	r := float32(math.Min(float64(cx), float64(cy))) / 6

	// Draw each of arc segments
	drawArc(cx, cy, 5*r, 5.2*r, float32(secondFraction), secondColor, gtx.Ops)
	drawArc(cx, cy, 4*r, 5*r, float32(minuteFraction), minuteColor, gtx.Ops)
	drawArc(cx, cy, 3*r, 4*r, float32(hourFraction), hourColor, gtx.Ops)
	drawArc(cx, cy, 2*r, 3*r, float32(dayFraction), dayColor, gtx.Ops)
	drawArc(cx, cy, r, 2*r, float32(yearFraction), yearColor, gtx.Ops)
	drawArc(cx, cy, 0, r, float32(centuryFraction), centuryColor, gtx.Ops)

	// After scheduling everything to be drawn, invalidate the whole screen so we
	// get drawn again soon.
	op.InvalidateOp{}.Add(gtx.Ops)
}

func drawArc(
	// x/y pixel coordinates of the center of the arc segment.
	cx, cy float32,
	// start/end radius of the arc segment. Must have r1 < r2.
	r1, r2 float32,
	// The fraction of the arc to render, starting with 0.0 is empty, 0.5 is a
	// quarter full filling the top right quadrant, and 1.0 is completely full.
	frac float32,
	col color.NRGBA,
	ops *op.Ops,
) {
	// Scale the [0,1] fraction to [0,-2π].
	angle := 2 * math.Pi * (-frac)
	// For sin/cos operations, normally angles have 0 being to the right and
	// proceed counter-clockwise.  We want 0 to be up, so we offset by π/4.
	angleFromUp := float64(angle) - math.Pi/2

	var path clip.Path
	path.Begin(ops)
	path.Move(f32.Pt(cx, cy-r1))                      // move to center, then straight up by r1
	path.LineTo(f32.Pt(cx, cy-r2))                    // move straight up to r2
	path.ArcTo(f32.Pt(cx, cy), f32.Pt(cx, cy), angle) // arc around by angle
	path.LineTo(f32.Pt(                               // move towards center from r2 to r1
		float32(math.Cos(angleFromUp))*r1+cx, // using trig here to get
		float32(math.Sin(angleFromUp))*r1+cy, // the position right
	))
	path.ArcTo(f32.Pt(cx, cy), f32.Pt(cx, cy), -angle) // arc back around by -angle to starting point
	defer clip.Outline{Path: path.End()}.Op().Push(ops).Pop()

	paint.ColorOp{Color: col}.Add(ops)
	paint.PaintOp{}.Add(ops)
}
