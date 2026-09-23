package notavalue

import (
	"math"
	"strconv"
)

// Format renders x as a string, keeping missing values and errors
// distinguishable:
//
//	Format(NaV)        → "NaV"
//	Format(math.NaN()) → "NaN"
//	Format(18.425)     → "18.425"
//	Format(math.Inf(1))→ "+Inf"
//
// Printing with %v or %g cannot do this: the standard library sees both
// sentinels as NaN-class and writes "NaN" for either, so the whole
// distinction this package maintains disappears on screen — in a log
// line, a table, or a failing test's message. Format is the way to keep
// it visible.
func Format(x float64) string {
	switch {
	case IsNaV(x):
		return "NaV"
	case math.IsNaN(x):
		return "NaN"
	case math.IsInf(x, 1):
		return "+Inf"
	case math.IsInf(x, -1):
		return "-Inf"
	}
	return strconv.FormatFloat(x, 'g', -1, 64)
}
