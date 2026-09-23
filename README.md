# notavalue

**A NaN-boxed sentinel for missing data in Go — and the arithmetic that
keeps computing when a value is not there.**

[![Go Reference](https://pkg.go.dev/badge/usefulrisk.com/notavalue.svg)](https://pkg.go.dev/usefulrisk.com/notavalue)

```go
import nav "usefulrisk.com/notavalue"

temperatures := []float64{18.2, 19.1, nav.NaV, 17.8, 18.6} // the sensor was offline

nav.Mean(temperatures)  // ≈ 18.425 — the mean of what was actually measured
nav.CountUsable(temperatures) // 4
nav.CountNaV(temperatures)    // 1
```

## The problem

Every series of measurements has holes. A sensor goes offline, a row is
missing from a join, a day was never recorded. Go gives you a
`[]float64` and no way to say "nothing here".

The usual answers all cost something:

| Approach | What it costs |
|---|---|
| Code the gap as `0` | Falsifies every sum and mean. A missing day becomes a day at zero. |
| Code it as `math.NaN()` | Honest, but contaminating: one gap in January turns the yearly mean into `NaN`. And a genuine computation error becomes indistinguishable from a gap. |
| Keep a parallel `[]bool` mask | Doubles the bookkeeping, costs memory, and gets forgotten at the first refactoring. |
| Use `[]*float64` or an option type | One pointer chase and one allocation per point, for a slice that was supposed to be flat. |

## The idea

`NaV` — "Not a Value" — is a quiet `NaN` with a tag bit set in its
mantissa. IEEE-754 leaves 51 bits free there, so the tag is free:

- **No memory overhead.** A `NaV` is an ordinary `float64`. Your
  `[]float64` stays a `[]float64`, with no companion mask.
- **Existing code keeps working.** `math.IsNaN(NaV)` is `true`, so any
  NaN-aware code you already have still guards correctly.
- **Only this package tells them apart.** `IsNaV` recognizes a missing
  value, `IsStdNaN` recognizes a real computation error.

## The three rules

1. **`NaV` never stops a computation.** It is skipped by aggregates and
   neutral in addition.
2. **`NaN` always propagates.** A broken computation must stay visible.
   When a `NaV` and a `NaN` meet, the `NaN` wins.
3. **Nothing left to compute on gives `NaV`.** An empty input, or one
   made of nothing but `NaV`.

```go
nav.Mean([]float64{1, 2, nav.NaV, 3})        // 2   — the gap is skipped
nav.Mean([]float64{1, 2, math.NaN(), 3})     // NaN — the error propagates
nav.Mean([]float64{nav.NaV, nav.NaV})        // NaV — nothing to say
```

The difference from `pandas` or `numpy` is that the two situations stay
distinguishable all the way through. A gap in the data and a division
by zero are not the same event, and a library about honest statistics
should not merge them.

## Arithmetic

Addition tolerates an absent term, the way a monthly mean tolerates a
missing day. A difference, a product and a quotient do not: they compare
or combine two values, and if one is unknown, so is the result.

```go
nav.Add(nav.NaV, 5)   // 5      NaV is the neutral element
nav.Sub(5, nav.NaV)   // NaV    returning 5 would read the gap as zero
nav.Mul(nav.NaV, 0)   // NaV    an unknown quantity of something stays unknown
nav.Div(nav.NaV, 0)   // NaV    nothing can be asserted, so nothing is
nav.Add(math.NaN(), nav.NaV) // NaN    the error wins
```

## What is in the box

**Predicates** — `IsNaV`, `IsStdNaN`

**Arithmetic** — `Add`, `Sub`, `Mul`, `Div`

**Inventories**, which never propagate anything because a tally cannot
break — `CountUsable`, `CountNaV`, `CountNaN`, `CountNonNaV`. The first
three partition the slice: their sum is always `len(xs)`.

**Aggregates** — `Sum`, `Mean`, `Min`, `Max`, `Bounds`, `Median`,
`Percentile`, `StdDev`

## Three decisions worth knowing before you rely on them

**`Median` returns a value that was actually observed.** It is defined
as `Percentile(xs, 50)`, so on an even count it returns the lower of the
two middle values — `20` for `[10 20 30 40]`, not `25`. Averaging two
observations manufactures a number nobody measured: on an ON/OFF signal
coded 0 and 1, the textbook median reports `0.5`, a state the equipment
never occupied. **This differs from the default of numpy, R and
pandas**, on even counts only, and it is intended.

**`StdDev` is the sample form**, with divisor `n-1`, computed over the
usable values.

**Infinities are ordinary numbers here.** `+Inf` is neither missing nor
erroneous, and it propagates through sums the way IEEE-754 prescribes.

## Cost

An aggregate reads its input once and allocates nothing. Measured on
100 000 `float64` with 5% of them missing, on an Apple M-series laptop:

| | Time | Allocations |
|---|---|---|
| `Sum`, `Mean`, `Bounds` | ~117 µs | 0 |
| `StdDev` | 201 µs | 0 |
| `Median` (copies and sorts) | 5.0 ms | 1 |

Scaling is linear up to ten million points — about 1.2 ns per point for
a mean. `Median` and `Percentile` are the two exceptions to the
zero-allocation rule: they sort a copy, so that your slice keeps its
order.

The benchmarks live in the repository, next to the alternatives that
were measured and rejected, and a test asserting that all of them return
the same result:

```
go test -run XXX -bench . -benchmem
```

## Install

```
go get usefulrisk.com/notavalue
```

`usefulrisk.com/notavalue` is the import path; the backing repository is
`github.com/fflamingodev/notavalue`. Import the vanity path, not the
repository URL.

Requires Go 1.21 or later. The package depends on nothing but the
standard library.

## Status

Pre-1.0. The semantics described above are settled and covered by tests,
but names and signatures may still move before `v1.0.0`.

## License

MIT. See [LICENSE](LICENSE).
