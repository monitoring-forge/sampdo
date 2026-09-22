# Radix sort benchmarks

Measured on Apple M3, darwin/arm64, Go 1.27.1, on 2026-09-22.

The baseline is `radixSort` from commit `e882f60`. Both implementations were
measured in the same test binary using a temporary side-by-side benchmark,
100 ms per sample and six samples per case. The table reports medians.
Input generation was excluded; each timed iteration copied the same deterministic
input into a reusable slice before sorting, matching the current
`BenchmarkSortDistributions` methodology. The temporary copy of the old sort
was removed after measurement. These are local measurements, not performance
guarantees for other CPUs or distributions.

| Elements | Distribution | Before (µs) | After (µs) | Speedup |
| ---: | --- | ---: | ---: | ---: |
| 1,000 | duplicates | 8.84 | 7.25 | 1.22× |
| 1,000 | random | 8.84 | 7.11 | 1.24× |
| 1,000 | wide | 9.82 | 4.74 | 2.07× |
| 1,000 | sorted | 0.51 | 0.26 | 1.97× |
| 1,000 | response_time | 26.25 | 7.83 | 3.35× |
| 10,000 | duplicates | 82.10 | 54.20 | 1.51× |
| 10,000 | random | 123.31 | 82.97 | 1.49× |
| 10,000 | wide | 94.58 | 65.91 | 1.43× |
| 10,000 | sorted | 5.83 | 3.38 | 1.72× |
| 10,000 | response_time | 277.15 | 72.95 | 3.80× |
| 100,000 | duplicates | 828.36 | 499.02 | 1.66× |
| 100,000 | random | 1,197.41 | 832.73 | 1.44× |
| 100,000 | wide | 900.32 | 634.49 | 1.42× |
| 100,000 | sorted | 52.58 | 26.94 | 1.95× |
| 100,000 | response_time | 2,780.05 | 657.68 | 4.23× |
| 1,000,000 | duplicates | 8,120.89 | 4,872.78 | 1.67× |
| 1,000,000 | random | 12,314.01 | 8,304.43 | 1.48× |
| 1,000,000 | wide | 9,431.60 | 6,346.96 | 1.49× |
| 1,000,000 | sorted | 599.04 | 334.09 | 1.79× |
| 1,000,000 | response_time | 27,772.10 | 6,137.81 | 4.52× |

Sorting uses one scratch allocation (802,816 B/op for 100,000 elements), as
before. Sorted inputs use zero allocations. The SIMD scan itself uses zero
allocations. Go may also grow its goroutine stack on the first call; the LSD
histograms occupy 96 KiB on 64-bit systems.

## `slices.Sort` vs `radixSort`

Measured on the same host and harness as above, after adding a reverse-order
fast path to `radixSort`. The table compares `slices.Sort` (Go 1.27's pdqsort)
with the package's `radixSort`. Each timed iteration copied the same
deterministic input into a reusable slice before sorting, so the copy cost is
included for both algorithms. Values are medians of six 100 ms samples.

| Elements | Distribution | `slices.Sort` (µs) | `radixSort` (µs) | Speedup |
| ---: | --- | ---: | ---: | ---: |
| 512 | duplicates | 6.63 | 4.35 | 1.52× |
| 512 | equal | 0.41 | 0.13 | 3.15× |
| 512 | random | 5.72 | 3.29 | 1.74× |
| 512 | response_time | 5.59 | 4.86 | 1.15× |
| 512 | reverse | 0.54 | 0.27 | 2.00× |
| 512 | reverse_duplicates | 2.07 | 0.27 | 7.67× |
| 512 | sorted | 0.41 | 0.13 | 3.15× |
| 512 | wide | 5.52 | 2.61 | 2.12× |
| 1,000 | duplicates | 8.80 | 7.24 | 1.22× |
| 1,000 | equal | 0.79 | 0.26 | 3.04× |
| 1,000 | random | 12.05 | 7.13 | 1.69× |
| 1,000 | response_time | 11.93 | 7.87 | 1.52× |
| 1,000 | reverse | 1.02 | 0.52 | 1.96× |
| 1,000 | reverse_duplicates | 3.12 | 0.52 | 6.00× |
| 1,000 | sorted | 0.79 | 0.26 | 3.04× |
| 1,000 | wide | 12.06 | 4.82 | 2.50× |
| 2,048 | duplicates | 19.89 | 13.14 | 1.51× |
| 2,048 | equal | 1.61 | 0.53 | 3.04× |
| 2,048 | random | 27.30 | 22.36 | 1.22× |
| 2,048 | response_time | 26.56 | 15.36 | 1.73× |
| 2,048 | reverse | 2.06 | 1.05 | 1.96× |
| 2,048 | reverse_duplicates | 6.91 | 1.05 | 6.58× |
| 2,048 | sorted | 1.62 | 0.52 | 3.12× |
| 2,048 | wide | 27.20 | 18.43 | 1.48× |
| 4,096 | duplicates | 36.15 | 24.99 | 1.45× |
| 4,096 | equal | 3.27 | 1.06 | 3.08× |
| 4,096 | random | 60.78 | 38.51 | 1.58× |
| 4,096 | response_time | 57.36 | 29.53 | 1.94× |
| 4,096 | reverse | 4.62 | 2.12 | 2.18× |
| 4,096 | reverse_duplicates | 14.53 | 2.10 | 6.92× |
| 4,096 | sorted | 3.35 | 1.06 | 3.16× |
| 4,096 | wide | 62.34 | 29.63 | 2.10× |
| 10,000 | duplicates | 93.18 | 54.81 | 1.70× |
| 10,000 | equal | 8.79 | 3.60 | 2.44× |
| 10,000 | random | 243.34 | 83.06 | 2.93× |
| 10,000 | response_time | 146.77 | 75.65 | 1.94× |
| 10,000 | reverse | 10.60 | 5.77 | 1.84× |
| 10,000 | reverse_duplicates | 32.21 | 6.03 | 5.34× |
| 10,000 | sorted | 9.15 | 3.55 | 2.58× |
| 10,000 | wide | 260.16 | 64.66 | 4.02× |
| 100,000 | duplicates | 1,465.99 | 509.70 | 2.88× |
| 100,000 | equal | 81.71 | 28.06 | 2.91× |
| 100,000 | random | 6,053.73 | 830.42 | 7.29× |
| 100,000 | response_time | 3,276.17 | 653.52 | 5.01× |
| 100,000 | reverse | 103.17 | 57.21 | 1.80× |
| 100,000 | reverse_duplicates | 305.79 | 56.72 | 5.39× |
| 100,000 | sorted | 81.46 | 27.63 | 2.95× |
| 100,000 | wide | 6,050.62 | 628.90 | 9.62× |
| 1,000,000 | duplicates | 15,794.87 | 4,853.88 | 3.25× |
| 1,000,000 | equal | 843.87 | 313.70 | 2.69× |
| 1,000,000 | random | 74,769.90 | 8,246.56 | 9.07× |
| 1,000,000 | response_time | 32,746.58 | 5,941.56 | 5.51× |
| 1,000,000 | reverse | 1,071.76 | 608.60 | 1.76× |
| 1,000,000 | reverse_duplicates | 3,118.17 | 610.98 | 5.10× |
| 1,000,000 | sorted | 843.92 | 313.21 | 2.69× |
| 1,000,000 | wide | 73,764.41 | 6,209.09 | 11.88× |

`slices.Sort` is in-place and allocation-free. `radixSort` uses one scratch
allocation for non-trivial inputs, but is substantially faster on random,
wide, duplicates, response-time, reverse, and reverse-duplicate distributions
once the element count exceeds a few thousand. For already-sorted or equal
inputs the radix scan overhead makes it slower than pdqsort's fast paths at
the smallest sizes, though it still wins by roughly 2.7–3× at 1,000,000
elements.

## SIMD scan in isolation

These measurements exclude the reset copy and sorting. Medians of three
100 ms samples, 100,000 elements:

| Distribution | Portable Go (µs) | ARM64 NEON (µs) | Speedup |
| --- | ---: | ---: | ---: |
| wide | 54.27 | 16.93 | 3.21× |
| sorted | 42.23 | 16.93 | 2.49× |

## Reproducing current measurements

```sh
go test -run='^$' -bench=BenchmarkSortDistributions -benchmem -benchtime=100ms -count=6 .
go test -run='^$' -bench=BenchmarkRadixScan -benchmem -benchtime=100ms -count=3 .
go test -run='^$' -bench=BenchmarkSampdoAll -benchmem -benchtime=100ms -count=3 .
```

When comparing revisions, use the same benchmark harness in both versions.
The previous `BenchmarkSortDistributions` excluded the reset copy and regenerated
input on every iteration, so its historical timings are not directly comparable
with the current harness.

## Validation

- `go test ./...`, `go test -race ./...`, and `go test -tags=purego ./...`.
- `golangci-lint run --timeout 5m ./...`.
- 30 seconds of `FuzzRadixSort`: 334,306 executions without a mismatch.
- Differential scan checks against portable Go, every adjacent inversion around
  SIMD boundaries, every non-sign bit, unaligned slices, and scalar tails.
- Sorting checks against `slices.Sort`, including subnormals, positive infinity,
  odd/even radix pass counts, both radix algorithms, and in-place aliasing.
- Linux/amd64 test binary cross-build and `go vet`; Linux/386 cross-build.

Only ARM64 was executed locally. The amd64 SSE2 implementation was cross-built
and vetted, but could not be executed on this host. The existing Linux/amd64 CI
test job exercises it. Other architectures and race builds use the portable scan.

## Descending-input follow-up

The follow-up to `7e6b298` detects a fully non-increasing input before the radix
scan and reverses it in place. The entire input is checked; descending endpoints
alone never cause an early return. This also supports descending inputs with
duplicate values, positive infinity, subnormal values, and zero.

Same Apple M3 / Go 1.27.1 environment, medians of three 100 ms samples, including
the same reset copy for both algorithms:

| Elements | slices.Sort (µs) | radixSort (µs) | Speedup vs slices.Sort |
| ---: | ---: | ---: | ---: |
| 512 | 0.53 | 0.27 | 1.98× |
| 1,000 | 1.01 | 0.51 | 1.97× |
| 2,048 | 2.15 | 1.05 | 2.04× |
| 4,096 | 4.07 | 2.08 | 1.96× |
| 10,000 | 10.56 | 5.73 | 1.84× |
| 100,000 | 104.93 | 57.15 | 1.84× |
| 1,000,000 | 1071.92 | 607.19 | 1.77× |

All descending cases allocate 0 B/op and 0 allocs/op. Before this follow-up,
100,000 descending elements took 593.38 µs and allocated 802,816 bytes; afterward
they take 57.15 µs with no allocation. Other existing distributions were also
remeasured at 1,000 through 1,000,000 elements, with median changes within ±5%.

Tests cover cutoff boundaries, duplicates, extreme values, in-place aliasing,
zero allocations on every reset-and-sort iteration, and every possible single
ascending pair within a 513-element descending input.

```sh
go test -run='^$' -bench='BenchmarkSortDistributions/(slices|radix)/.*/reverse$' -benchmem -benchtime=100ms -count=3 .
```
