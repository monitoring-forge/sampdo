# sampdo

`sampdo` は Go 用の軽量な数値集計ライブラリです。大量の `float64` サンプルに対して、ソート済みの状態から最小値・最大値・平均値・中央値・任意のパーセンタイルを効率的に取得できます。非負数に対しては基数ソートを使い、比較ソートより高速に集計できます。

## 特徴

- **ソートベースの集計**: サンプルを 1 回ソートするだけで、複数の統計量を取得できます。
- **基数ソート対応**: 非負数の場合、`slices.Sort` よりも高速な基数ソートを使用します。
- **NaN を排除**: `Append` 時に `NaN` を弾き、集計結果を安定させます。
- **イミュータブルなソート結果**: `Sorted()` 後に元のデータを変更しても、返された `*Sorted` は影響を受けません。
- **無限大に対応**: `+Inf`、`-Inf` を含むデータのパーセンタイル計算にも対応しています。

## インストール

```bash
go get github.com/monitoring-forge/sampdo
```

## 使い方

```go
package main

import (
	"fmt"
	"log"

	"github.com/monitoring-forge/sampdo"
)

func main() {
	s := sampdo.New()

	// サンプルを追加（可変長引数）
	if err := s.Append(8.0, 2.0, 10.0, 4.0, 6.0, 3.0, 7.0, 1.0, 9.0, 5.0); err != nil {
		log.Fatal(err)
	}

	sorted, err := s.Sorted()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Count  :", sorted.Count())       // サンプル数
	fmt.Println("Min    :", sorted.Min())         // 最小値
	fmt.Println("Max    :", sorted.Max())         // 最大値
	fmt.Println("Mean   :", sorted.Mean())        // 平均値
	fmt.Println("Median :", sorted.Median())      // 中央値
	fmt.Println("P75    :", sorted.Percentile(75)) // 75 パーセンタイル
	fmt.Println("P90    :", sorted.Percentile(90)) // 90 パーセンタイル
	fmt.Println("P95    :", sorted.Percentile(95)) // 95 パーセンタイル
	fmt.Println("P99    :", sorted.Percentile(99)) // 99 パーセンタイル
}
```

### 出力例

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

## 出力される数字の意味

| メソッド | 説明 |
| --- | --- |
| `Count()` | 追加されたサンプルの個数を返します。 |
| `Min()` | 最小値を返します。サンプルが空の場合は `0` を返します。 |
| `Max()` | 最大値を返します。サンプルが空の場合は `0` を返します。 |
| `Mean()` | 平均値（算術平均）を返します。サンプルが空の場合は `0` を返します。入力された順序で合計を計算します。 |
| `Median()` | 中央値（50 パーセンタイル）を返します。要素数が偶数の場合は中央 2 値の平均を返します。 |
| `Percentile(p)` | `p` パーセンタイル（0〜100）を線形補間で返します。範囲外の `p` には `0` を返します。 |

### パーセンタイルの計算方法

`Percentile(p)` は以下の式で位置を求め、隣接する 2 点間を線形補間します。

```
k = (p / 100) * (n - 1)
```

`n` はサンプル数です。例えば 10 サンプルの P75 は `k = 6.75` となり、7 番目と 8 番目の値の間を 0.75 で補間します。

## オプション

`New` に以下のオプションを渡せます。

```go
// 初期キャパシティを指定（デフォルトは 128）
s := sampdo.New(sampdo.WithInitialCapacity(1024))

// 常に slices.Sort を使う（非負数でも基数ソートを使わない）
s := sampdo.New(sampdo.WithPreferSlicesSort(true))
```

## 制限・注意事項

- **NaN は追加不可**: `Append` に `NaN` を含めるとエラーになります。
- **負のゼロは比較ソートにフォールバック**: `-0` を含めると負の数と同様に `slices.Sort` が使用されます。
- **空の集計はエラー**: 1 つも `Append` せずに `Sorted()` を呼ぶとエラーを返します。
- **パーセンタイルは 0〜100 のみ**: 範囲外の値に対しては `0` を返します。
- **`Sorted()` は 1 回だけ呼ぶ想定**: 同じ `*Sampdo` に対して `Sorted()` 後に `Append` して再集計できますが、ソート済みスライスをクローンするためコストが発生します。

## ベンチマーク

以下のコマンドでベンチマークを実行できます。

```bash
go test -bench BenchmarkSampdoAll -benchmem -benchtime=100ms -count 10 -run '^$' .
```

## ライセンス

MIT License

Copyright (c) 2026 Masahiro Nagano

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

