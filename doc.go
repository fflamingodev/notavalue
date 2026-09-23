// Package notavalue provides NaV, a NaN-boxed sentinel for missing
// data, and the arithmetic and statistics that go with it.
//
// It answers a question every series of measurements raises: what
// should happen to a computation when a value is simply not there? A
// sensor was offline, a row was absent from a join, a day was never
// recorded. Coding that absence as 0 falsifies the sums; coding it as
// math.NaN() makes every later result an error; keeping a parallel
// array of booleans costs memory and is forgotten at the first
// refactoring.
//
// # Two kinds of non-numbers
//
// The package distinguishes two kinds of non-numeric float64, and
// treats them in opposite ways:
//
//  1. NaV ("Not a Value") marks a missing observation. A missing value
//     never stops a computation: NaV is skipped by aggregates and is
//     neutral in addition.
//
//  2. NaN marks a computation error (0/0, log(-1), Inf-Inf, ...). It
//     always propagates, so the error stays visible in the result. When
//     a NaV and a NaN meet, the NaN wins.
//
//  3. When there is nothing left to compute on — an empty input, or one
//     made of nothing but NaV — the result is NaV.
//
// The rationale: one missing day must not turn a monthly mean into an
// error, whereas a broken computation must never be silently hidden.
//
//	Mean([]float64{1, 2, NaV, 3})        // 2   — the missing value is skipped
//	Mean([]float64{1, 2, math.NaN(), 3}) // NaN — the error propagates
//	Mean([]float64{NaV, NaV})            // NaV — nothing to say
//
// # How it works: NaN boxing
//
// IEEE-754 leaves 51 free bits in the mantissa of a quiet NaN. NaV is a
// quiet NaN with one of them set as a tag. The consequences are what
// make the idea practical:
//
//   - no memory overhead at all: a NaV is an ordinary float64, and a
//     []float64 needs no companion mask;
//   - math.IsNaN reports a NaV as NaN, so existing NaN-aware code keeps
//     working untouched;
//   - only IsNaV tells the two apart, and IsStdNaN isolates real
//     computation errors.
//
// The remaining free bits are room for future distinctions — missing,
// interpolated, rejected — without breaking compatibility.
//
// # What the aggregates cost
//
// Skipping missing values must not make the package slow, because these
// functions are meant to run over long series and over many of them.
// So, as a rule, an aggregate reads its input once and allocates
// nothing. Sum, Mean, Min, Max and Bounds all work that way, out of a
// single shared traversal.
//
// Two exceptions, both deliberate:
//
//   - Median and Percentile must sort, so they copy the usable values
//     first. The caller's slice keeps its order, which matters: in a
//     time series, order carries meaning.
//   - StdDev reads the input twice, since the mean must be known before
//     the deviations can be squared. It still copies nothing. The
//     one-pass alternatives were measured and rejected; the reasons are
//     in its documentation.
//
// Every one of these choices is backed by a benchmark kept in the
// repository, in aggregates_bench_test.go, together with the rejected
// alternatives and a test asserting that they all return the same
// result. Run them with:
//
//	go test -run XXX -bench . -benchmem
//
// # Importing
//
// The import path spells out what the package is about; a short alias
// keeps the calls readable:
//
//	import nav "usefulrisk.com/notavalue"
//
//	x := nav.NaV
package notavalue
