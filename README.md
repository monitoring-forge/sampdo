# sampdo

`sampdo` is a lightweight statistical aggregation library for Go. It collects a large number of `float64` samples and efficiently computes minimum, maximum, mean, median, and arbitrary percentiles from a sorted view. For non-negative values it uses radix sort, which is faster than a comparison sort.

`sampdo` は Go 用の軽量な数値集計ライブラリです。大量の `float64` サンプルに対して、ソート済みの状態から最小値・最大値・平均値・中央値・任意のパーセンタイルを効率的に取得できます。非負数に対しては基数ソートを使い、比較ソートより高速に集計できます。

## Features / 特徴

- **Sort-based aggregation / ソートベースの集計**: Compute several statistics from a single sort. サンプルを 1 回ソートするだけで複数の統計量を取得できます。
- **Radix sort for non-negative numbers / 非負数向け基数ソート**: Falls back to `slices.Sort` only when negative or `-0` values are present. 負数や `-0` がある場合のみ `slices.Sort` にフォールバックします。
- **NaN rejection / NaN を排除**: `Append` rejects `NaN` so results stay stable. `Append` は `NaN` を弾き、集計結果を安定させます。
- **Immutable sorted view / イミュータブルなソート結果**: Appending after `Sorted()` does not mutate the returned `*Sorted`. `Sorted()` 後に追加しても返された `*Sorted` は影響を受けません。
- **Infinity support / 無限大対応**: Handles `+Inf` and `-Inf` in percentile calculations. `+Inf`、`-Inf` を含むデータのパーセンタイル計算にも対応しています。

## Installation / インストール

```bash
go get github.com/monitoring-forge/sampdo
```

## Usage / 使い方

```go
package main

import (
	"fmt"
	"log"

	"github.com/monitoring-forge/sampdo"
)

func main() {
	s := sampdo.New()

	// Add samples (variadic) / サンプルを追加（可変長引数）
	if err := s.Append(8.0, 2.0, 10.0, 4.0, 6.0, 3.0, 7.0, 1.0, 9.0, 5.0); err != nil {
		log.Fatal(err)
	}

	sorted, err := s.Sorted()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Count  :", sorted.Count())        // number of samples / サンプル数
	fmt.Println("Min    :", sorted.Min())          // minimum value / 最小値
	fmt.Println("Max    :", sorted.Max())          // maximum value / 最大値
	fmt.Println("Mean   :", sorted.Mean())         // arithmetic mean / 平均値
	fmt.Println("Median :", sorted.Median())       // median value / 中央値
	fmt.Println("P75    :", sorted.Percentile(75)) // 75th percentile / 75 パーセンタイル
	fmt.Println("P90    :", sorted.Percentile(90)) // 90th percentile / 90 パーセンタイル
	fmt.Println("P95    :", sorted.Percentile(95)) // 95th percentile / 95 パーセンタイル
	fmt.Println("P99    :", sorted.Percentile(99)) // 99th percentile / 99 パーセンタイル
}
```

### Example output / 出力例

```
Count  : 10
Min    : 1
Max    : 10
Mean   : 5.5
Median : 5.5
P75    : 7.75
P90    : 9.1
P95    : 9.55
P99    : 9.91
```

## What the numbers mean / 出力される数字の意味

| Method | English | 日本語 |
| --- | --- | --- |
| `Count()` | Number of appended samples. | 追加されたサンプルの個数。 |
| `Min()` | Smallest value; returns `0` when empty. | 最小値。空の場合は `0`。 |
| `Max()` | Largest value; returns `0` when empty. | 最大値。空の場合は `0`。 |
| `Mean()` | Arithmetic mean; sums in input order. Returns `0` when empty. | 算術平均。入力順に合計し、空の場合は `0`。 |
| `Median()` | 50th percentile; averages the two middle values for an even count. | 50 パーセンタイル。要素数が偶数の場合は中央 2 値の平均。 |
| `Percentile(p)` | Returns the `p`th percentile (0–100) using linear interpolation; returns `0` for out-of-range `p`. | `p` パーセンタイル（0〜100）を線形補間で返す。範囲外は `0`。 |

### How percentiles are calculated / パーセンタイルの計算方法

`Percentile(p)` computes the position `k` with the formula below and linearly interpolates between the two surrounding points.

`Percentile(p)` は以下の式で位置 `k` を求め、隣接する 2 点間を線形補間します。

```
k = (p / 100) * (n - 1)
```

`n` is the sample count. For example, with 10 samples, P75 yields `k = 6.75`, so the result is interpolated 75% of the way between the 7th and 8th values.

`n` はサンプル数です。例えば 10 サンプルの P75 は `k = 6.75` となり、7 番目と 8 番目の値の間を 0.75 で補間します。

## Options / オプション

Pass these options to `New`:

`New` に以下のオプションを渡せます。

```go
// Set the initial capacity (default is 128) / 初期キャパシティを指定（デフォルトは 128）
s := sampdo.New(sampdo.WithInitialCapacity(1024))

// Always use slices.Sort, even for non-negative data / 常に slices.Sort を使う
s := sampdo.New(sampdo.WithPreferSlicesSort(true))
```

## Limitations and notes / 制限・注意事項

- **NaN is not allowed / NaN は追加不可**: `Append` returns an error if any value is `NaN`. `NaN` を含めると `Append` はエラーを返します。
- **Negative zero falls back to comparison sort / 負のゼロは比較ソートにフォールバック**: Including `-0` causes the library to use `slices.Sort`. `-0` を含めると `slices.Sort` が使用されます。
- **Empty input is an error / 空の集計はエラー**: Calling `Sorted()` without any appended points returns an error. 1 つも `Append` せずに `Sorted()` を呼ぶとエラーになります。
- **Percentiles are limited to 0–100 / パーセンタイルは 0〜100 のみ**: Out-of-range percentiles return `0`. 範囲外の値に対しては `0` を返します。
- **Reuse after `Sorted()` has a cost / `Sorted()` 後の再利用にコスト**: You can keep appending after `Sorted()`, but the sorted slice is cloned to keep the previous `*Sorted` immutable. `Sorted()` 後に追加できますが、イミュータブル性を保つためソート済みスライスのクローンが発生します。

## Benchmark / ベンチマーク

Run benchmarks with:

以下のコマンドでベンチマークを実行できます。

```bash
go test -bench BenchmarkSampdoAll -benchmem -benchtime=100ms -count 10 -run '^$' .
```

## License / ライセンス

This project is licensed under the [MIT License](LICENSE).

本プロジェクトは [MIT ライセンス](LICENSE) の下で提供されています。

