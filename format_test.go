package notavalue

import (
	"fmt"
	"math"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{NaV, "NaV"},
		{math.NaN(), "NaN"},
		{0, "0"},
		{18.425, "18.425"},
		{-3, "-3"},
		{math.Inf(1), "+Inf"},
		{math.Inf(-1), "-Inf"},
	}
	for _, c := range cases {
		if got := Format(c.in); got != c.want {
			t.Errorf("Format(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The reason Format exists: the standard library prints both sentinels
// the same way, and the distinction vanishes on screen.
func TestFormatSucceedsWhereFmtCannot(t *testing.T) {
	if fmt.Sprintf("%v", NaV) != fmt.Sprintf("%v", math.NaN()) {
		t.Skip("this Go version already prints NaV and NaN differently")
	}
	if Format(NaV) == Format(math.NaN()) {
		t.Error("Format renders NaV and NaN identically, which defeats its purpose")
	}
}

func TestCoalesce(t *testing.T) {
	cases := []struct {
		name  string
		in    []float64
		want  want
		value float64
	}{
		{"the first value is usable", []float64{5, 7}, wantNum, 5},
		{"missing values are passed over", []float64{NaV, NaV, 5, 7}, wantNum, 5},
		{"zero is a value", []float64{NaV, 0, 7}, wantNum, 0},
		{"nothing usable", []float64{NaV, NaV}, wantNaV, 0},
		{"no argument at all", nil, wantNaV, 0},
		// Rule 2: an error is not stepped around to reach the next candidate.
		{"an error is returned, not skipped", []float64{math.NaN(), 5}, wantNaN, 0},
		{"an error after a gap", []float64{NaV, math.NaN(), 5}, wantNaN, 0},
		// ... but an error that comes after a usable value is never reached.
		{"an error behind a usable value is not seen", []float64{5, math.NaN()}, wantNum, 5},
	}
	for _, c := range cases {
		check(t, "Coalesce "+c.name, Coalesce(c.in...), c.want, c.value)
	}
}

// The typical use: a primary sensor, its backup, then a modeled value.
func TestCoalescePicksTheFirstAvailableSource(t *testing.T) {
	primary, backup, modeled := NaV, NaV, 19.0
	check(t, "Coalesce(primary, backup, modeled)",
		Coalesce(primary, backup, modeled), wantNum, 19)

	backup = 18.5
	check(t, "Coalesce with the backup online",
		Coalesce(primary, backup, modeled), wantNum, 18.5)
}
