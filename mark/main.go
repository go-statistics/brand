// Command mark writes the go-statistics mark, svg/color/go-statistics.svg:
//
//	cd mark && go run . > ../svg/color/go-statistics.svg
//
// The mark is a normal density over a baseline, with the two lines at one
// standard deviation either side of the mean that bound its central 68 %.
// The curve is not drawn by hand: it is the density itself, sampled at every
// half standard deviation and joined by cubic Bézier segments whose control
// points come from the density's exact derivative (cubic Hermite
// interpolation), so the path is as smooth as the function it draws.
//
// The rest is the family grammar every go-* mark shares: a 256-unit rounded
// square (rx 56) filled with a diagonal two-stop gradient, a white line glyph
// at stroke-width 13 with round caps and joins. logos/brandgen renders the
// format matrix from the file this writes; the kit workflow runs this again
// and fails if a byte of svg/color differs.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

const (
	light = "#F2B36B" // the gradient's light stop
	dark  = "#9E520F" // its dark stop

	baseline = 196.0 // y of the axis
	peak     = 132.0 // height of the density at the mean
	mean     = 128.0 // x of the mean: the middle of the square
	sigma    = 28.0  // one standard deviation, in user units
	reach    = 3.0   // the curve spans mean ± reach·sigma
	step     = 0.5   // sampling step, in standard deviations
)

// floor is the density's value at the ends of the span, relative to its mode.
// It is taken off so that the curve lands exactly on the axis instead of
// stopping a stroke's width above it.
var floor = math.Exp(-reach * reach / 2)

// y is the density, scaled into the square: the axis at baseline, the mode
// peak units above it, the ends of the span on the axis.
func y(x float64) float64 {
	z := (x - mean) / sigma
	return baseline - peak*(math.Exp(-z*z/2)-floor)/(1-floor)
}

// dy is y's exact derivative with respect to x.
func dy(x float64) float64 {
	z := (x - mean) / sigma
	return peak / (1 - floor) * math.Exp(-z*z/2) * z / sigma
}

// num prints a coordinate with at most one decimal and no trailing zero.
func num(v float64) string {
	s := fmt.Sprintf("%.1f", math.Round(v*10)/10)
	s = strings.TrimSuffix(s, ".0")
	if s == "-0" {
		s = "0"
	}
	return s
}

func pt(x, y float64) string { return num(x) + " " + num(y) }

// curve is the density from mean-reach·sigma to mean+reach·sigma as one path
// of cubic segments: for a segment [x0, x1] of width h, the control points
// are a third of the way along each end's tangent.
func curve() string {
	var b strings.Builder
	n := int(math.Round(2 * reach / step))
	x0 := mean - reach*sigma
	b.WriteString("M" + pt(x0, y(x0)))
	for i := 0; i < n; i++ {
		a := mean + (-reach+float64(i)*step)*sigma
		c := a + step*sigma
		h := c - a
		b.WriteString("C" + pt(a+h/3, y(a)+h/3*dy(a)) + " " +
			pt(c-h/3, y(c)-h/3*dy(c)) + " " + pt(c, y(c)))
	}
	return b.String()
}

func main() {
	lo, hi := mean-sigma, mean+sigma
	axis := fmt.Sprintf("M%s H%s", pt(36, baseline), num(220))
	bounds := fmt.Sprintf(`<path d="M%s V%s"/><path d="M%s V%s"/>`,
		pt(lo, baseline), num(y(lo)), pt(hi, baseline), num(y(hi)))

	fmt.Fprintf(os.Stdout, `<svg viewBox="0 0 256 256" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="go-statistics">
  <defs><linearGradient id="statgrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient></defs>
  <rect width="256" height="256" rx="56" fill="url(#statgrad)"/>
  <g fill="none" stroke="#FFFFFF" stroke-width="13" stroke-linecap="round" stroke-linejoin="round">
    <path d="%s"/>
    <path d="%s"/>
    %s
  </g>
</svg>
`, light, dark, curve(), axis, bounds)
}
