# set

使用闭区间存储整数集合的 Go 模块，适用于端口、IPv4 整数地址及其他整数范围。模块路径为 `github.com/gatework/set`，需要 Go 1.27.0 或更高版本，仅依赖标准库。

例如，整个 `uint64` 值域只占一个区间。运算量取决于区间数量，不逐个存储或遍历范围内的整数。

```go
package main

import (
	"fmt"

	"github.com/gatework/set"
)

func main() {
	var ports set.Set[uint16]
	ports.Add(80, 80)
	ports.Add(443, 443)
	ports.Add(8080, 8082)
	ports.Add(8083, 8085)

	allowed := set.New(set.Range[uint16]{Min: 440, Max: 450})
	fmt.Println(&ports)                      // {[80,80], [443,443], [8080,8085]}
	fmt.Println(ports.Contains(443))         // true
	fmt.Println(ports.Intersection(allowed)) // {[443,443]}
	fmt.Println(ports.Len())                 // 3 个区间
}
```

版本：[`v0.1.0`](https://github.com/gatework/set/releases/tag/v0.1.0)。公开模块，可直接按版本安装：

```sh
go get github.com/gatework/set@v0.1.0
```

在模块目录内可执行 `go doc .` 查看 API，执行 `make check` 运行检查。

## 集合契约

- `Integer` 支持 `int`、`int8`、`int16`、`int32`、`int64`、`uint`、`uint8`、`uint16`、`uint32`、`uint64`、`uintptr`，以及以这些类型为底层类型的自定义类型。`byte`、`rune` 别名也可直接使用。
- `Range[T]{Min, Max}` 的两个端点均包含在范围内。`New` 和 `Add` 自动交换反向端点，排序并合并重叠或相邻区间。
- `Set[T]` 零值为空集，`Range[T]{}` 是单点零。`nil *Set[T]` 接收者和参数不受支持。
- `Len` 返回规范化后的区间数量。空集的 `Min`、`Max` 返回类型零值和 `false`。
- 相等表示覆盖的整数完全相同。空集是每个集合的子集；非空集严格包含空集。
- 集合本身不负责 IP 文本解析、序列化、差集或补集。

## API

所有集合方法均使用 `*Set[T]` 接收者，两个集合参与运算时需要相同的 `T`。

| 接口 | 行为 |
| --- | --- |
| `New[T](ranges ...Range[T]) *Set[T]` | 复制输入并规范化；无参数时构造空集 |
| `Add(lo, hi T)` | 将一个闭区间并入接收者 |
| `Merge(other *Set[T])` | 将另一个集合并入接收者；可与自身合并 |
| `Clone() *Set[T]` | 返回拥有独立底层切片的副本 |
| `Union(other *Set[T]) *Set[T]` | 返回两个输入的并集 |
| `Intersection(other *Set[T]) *Set[T]` | 返回两个输入的交集 |
| `Contains(value T) bool` | 判断一个整数是否属于集合 |
| `Equal(other *Set[T]) bool` | 判断覆盖范围是否相同 |
| `IsSubset`、`IsSuperset` | 判断子集、超集，包含相等情况 |
| `IsProperSubset`、`IsProperSuperset` | 判断严格子集、严格超集 |
| `Compare(other *Set[T]) Relation` | 相等、严格超集、严格子集或其他 |
| `Len() int`、`IsEmpty() bool` | 区间数量及空集判断 |
| `Min() (T, bool)`、`Max() (T, bool)` | 最小值、最大值及是否存在 |
| `Ranges() []Range[T]` | 返回按下界排序的区间副本；空集可能返回 `nil` |
| `All() iter.Seq[Range[T]]` | 返回调用时刻的区间快照迭代器 |
| `String() string` | 返回 `{[1,4], [7,10]}` 形式的十进制描述；空集为 `{}` |

`Relation` 的零值是 `RelationUnknown`，`Compare` 只返回以下四种结果：

| 常量 | 接收者相对参数的关系 |
| --- | --- |
| `RelationEqual` | 相等 |
| `RelationSuperset` | 严格超集 |
| `RelationSubset` | 严格子集 |
| `RelationOther` | 互不包含，包括部分重叠和完全不相交 |

这些关系不构成供排序使用的全序。

## 修改、复制和遍历

`Add`、`Merge` 修改接收者，其余运算不修改输入。写入先生成新切片，再替换接收者状态；已经提供给其他集合或迭代器的底层切片保持只读。

```go
original := set.New(set.Range[int]{Min: 1, Max: 4})
cloned := original.Clone() // 显式复制底层存储
valueCopy := *original    // 值复制同样不会受原集合后续修改影响
alias := original        // 指针复制仍指向同一个对象

original.Add(5, 8)
fmt.Println(cloned)     // {[1,4]}
fmt.Println(&valueCopy) // {[1,4]}
fmt.Println(alias)      // {[1,8]}
```

`New` 不保留调用方切片，修改 `Ranges()` 返回的切片也不会影响集合。`Union`、`Intersection` 返回独立结果，包括输入或输出为空集的情况。

`All()` 遍历区间，不枚举区间内每个整数。快照在调用 `All` 时确定，可重复遍历，并支持 `break`。之后修改原集合，或在同一 goroutine 的遍历回调内修改原集合，都不会改变该快照。

```go
numbers := set.New(set.Range[int]{Min: -2, Max: 0}, set.Range[int]{Min: 4, Max: 5})
snapshot := numbers.All()
numbers.Add(1, 3)

for r := range snapshot {
	fmt.Println(r.Min, r.Max) // -2 0，然后 4 5
}
fmt.Println(numbers) // {[-2,5]}
```

没有写入时，同一集合可被并发读取。同一对象的并发读写、并发写入以及写入期间获取快照需要调用方同步；模块不包含锁。安全交接给其他 goroutine 的已有快照和值副本不会受原集合后续写入影响。

## 算法

令输入区间数为 `n`、`m`，交集结果区间数为 `k`：

| 操作 | 时间 | 额外空间 |
| --- | --- | --- |
| 批量构造 `New` | `O(n log n)` | `O(n)` |
| `Add` | `O(n)` | `O(n)` |
| `Merge`、`Union` | `O(n+m)` | `O(n+m)` |
| `Intersection` | `O(n+m)` | `O(k)` |
| 子集、超集和关系判断 | `O(n+m)` | `O(1)` |
| `Contains` | `O(log n)` | `O(1)` |
| `Len`、`IsEmpty`、`Min`、`Max` | `O(1)` | `O(1)` |
| `Clone`、`Ranges` | `O(n)` | `O(n)` |
| 获取 `All` 快照 | `O(1)` | `O(1)` |

批量输入优先使用 `New(ranges...)`，避免循环添加大量分散区间产生反复复制。相邻判断在证明端点运算安全后才执行减一，不使用可能溢出的端点相减排序。

## 与 Perl 参考实现的对应关系

算法语义参考 `Firewall/Utils/Set.pm` 及其 `set.t`、`set-properties.t`、`boundaries.t` 测试；Go 模块和测试均不依赖 Perl。

| Perl | Go |
| --- | --- |
| 空集、端点或对象复制构造 | 零值、`New`、`Clone` |
| `mins`、`maxs` 平行数组 | 封装的区间切片，通过 `Ranges` 或 `All` 读取 |
| `addToSet` | `Add`，始终维护规范化，不保留未合并的相邻片段 |
| `mergeToSet(min, max)`、`mergeToSet(set)` | `Add`、`Merge` |
| `interSet` | `Intersection`；单区间相交作为内部辅助函数 |
| `isEqual`、`notEqual` | `Equal`、`!Equal`；按实际覆盖范围比较 |
| 包含、被包含及严格关系 | `IsSuperset`、`IsSubset` 及 `IsProper*` |
| 四种字符串关系 | `Relation` 类型化常量 |
| `length`、`min`、`max` | `Len`、`Min`、`Max`；空值使用 `(T, bool)` |
| `dump` | `String` 和 `fmt.Stringer`，由调用方决定是否输出 |

Go 版本还明确支持有符号整数和完整的 `uint64` 值域。

## 开发与验证

```sh
make check # gofmt 检查、go test -race ./...、go vet ./...
make fuzz  # 30 秒 FuzzSet，使用独立的逐点模型
```

Makefile 设置 `GOWORK=off`，可在周边存在其他模块时独立验证。也可以直接执行：

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go test -run '^$' -fuzz '^FuzzSet$' -fuzztime=30s .
```

测试覆盖六点域的全部 4096 个集合对、各整数类型的极值、构造和返回切片隔离、值复制、空结果、自身合并、迭代快照与提前退出。Fuzz 使用有符号八位整数完整值域的位图校验实际覆盖范围，验证标准集合关系及复制隔离。

修改后运行 `make fmt` 和 `make check`；涉及集合算法或存储所有权的修改还应运行 `make fuzz`。如已安装支持 Go 1.27 的 golangci-lint v2，可额外运行 `make lint`。

## 发布

CI 对 `main`、Pull Request、手动触发和 Release 调用运行格式、工作流校验、race、vet、fuzz、lint 和可达漏洞检查。非 Pull Request 工作流通过后，将同一提交的源码归档和 `SHA256SUMS` 上传为 `source-<commit>` artifact。

发布流程：

1. 在 [CHANGELOG](CHANGELOG.md) 添加 `## vX.Y.Z — 日期` 条目，将修改合入 `main`。
2. 为 `main` 中的提交创建并推送 `vX.Y.Z` 标签。Release 工作流支持稳定版本标签，不接受预发布后缀。
3. Release 校验标签和版本说明，调用 CI，下载本次运行生成的归档，并核对校验和及源码提交。
4. 工作流创建或恢复草稿，上传 `set-source.tar.gz`、`SHA256SUMS` 并下载核对后发布。已有附件必须逐字节一致，不会被覆盖。

如果发布阶段失败，优先在 Actions 中重跑失败作业，以复用已经完成的 CI artifact。也可以按标签手动触发完整流程：

```sh
gh workflow run release.yml --ref v0.1.0
```

手动触发时必须选择版本标签。已发布版本重跑时只核对已有附件；缺失或不一致会失败。库不提供可执行二进制，消费者使用 `go get github.com/gatework/set@vX.Y.Z` 获取模块。
