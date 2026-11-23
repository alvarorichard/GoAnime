<p align="center">
  <img src="https://user-images.githubusercontent.com/65846651/229838021-741ff719-8c99-45f6-88d2-1a32927bd863.png">
</p>

# g - Functional programming framework for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/enetx/g.svg)](https://pkg.go.dev/github.com/enetx/g)
[![Coverage Status](https://coveralls.io/repos/github/enetx/g/badge.svg?branch=main&service=github)](https://coveralls.io/github/enetx/g?branch=main)
[![Go](https://github.com/enetx/g/actions/workflows/go.yml/badge.svg)](https://github.com/enetx/g/actions/workflows/go.yml)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/enetx/g)

```bash
go get github.com/enetx/g
```

Requires **Go 1.27+** (generic methods: `Map`, `Then`, `Fold`, `Zip` and friends change element types right in the chain — no more package-level `Transform*` workarounds).

Designed to be cheap to compile: the type graph is deliberately de-welded, so a
package that only uses `g.String` compiles ~1.5k functions instead of ~33k, and
heavyweight areas (filesystem, regex) live in subpackages (`g/fs`, `g/rx`).
The only non-stdlib dependency is `golang.org/x/text`.

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/enetx/g"
)

func main() {
    // Slice with functional operations
    g.SliceOf(1, 2, 3, 4, 5).
        Iter().
        Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) int { return x * x }).
        Collect().
        Slice().
        Println() // Slice[4, 16]

    // Safe map access with Option
    m := g.NewMap[string, int]()
    m.Insert("key", 42)

    value := m.Get("key").UnwrapOr(0)
    fmt.Println(value) // 42

    // Result for error handling
    result := g.String("123").TryInt()
    if result.IsOk() {
        fmt.Println(result.Unwrap()) // 123
    }
}
```

---

## Navigation

| Core | Collections | Sync | Types |
|:----:|:-----------:|:----:|:-----:|
| [Option](#option) | [Slice](#slice) | [Mutex](#mutex) | [String](#string) |
| [Result](#result) | [Map](#map) | [RwLock](#rwlock) | [Int/Float](#int--float) |
| [Iterators](#iterators) | [Set](#set) | [Pool](#pool) | [Bytes](#bytes) |
| | [Deque](#deque) | | [Regex (g/rx)](#regular-expressions--grx) |
| | [Heap](#heap) | | [File/Dir (g/fs)](#file--directory) |

---

## Option

Safe nullable values: `Some(value)` or `None`.

```go
opt := g.Some(42)           // Some(42)
none := g.None[int]()       // None

opt.IsSome()                // true
opt.Unwrap()                // 42 (panics if None)
opt.UnwrapOr(0)             // 42 (returns default if None)
opt.UnwrapOrDefault()       // 42 (returns zero value if None)

// From map lookup
v, ok := myMap[key]
opt := g.OptionOf(v, ok)

// Chaining
g.Some(5).Then(func(x int) g.Option[int] {
    return g.Some(x * 2)
}) // Some(10)
```

<details>
<summary><b>All Option Methods</b></summary>

| Method | Description |
|--------|-------------|
| `IsSome()` / `IsNone()` | Presence checks |
| `IsSomeAnd(pred)` / `IsNoneOr(pred)` | Presence combined with a predicate |
| `Some()` | Returns value (zero value if None — check first) |
| `Unwrap()` / `Expect(msg)` | Returns value, panics if None |
| `UnwrapOr(default)` / `UnwrapOrDefault()` | Returns value or fallback |
| `Map(fn)` | Transforms the value, may change its type |
| `MapOr(def, fn)` / `MapOrElse(defFn, fn)` | Map or fall back to a default |
| `Then(fn)` | Chains a fallible transformation (and_then) |
| `Filter(pred)` | Returns None if predicate fails |
| `Or(other)` / `OrElse(fn)` | Fallback Option |
| `Inspect(fn)` | Side effect on Some, returns self |
| `g.OkOr(opt, err)` / `g.OkOrElse(opt, fn)` | Converts to Result (free functions) |
| `Take()` / `Replace(v)` / `Insert(v)` | In-place mutation |
| `Option()` | Returns (value, bool) |

JSON: `Some(v)` ⇄ value, `None` ⇄ `null` (encoding/json/v2 native, v1 compatible).

</details>

---

## Result

Success (`Ok`) or failure (`Err`) with error.

```go
ok := g.Ok(42)                              // Ok(42)
err := g.Err[int](errors.New("failed"))     // Err

ok.IsOk()                   // true
ok.Unwrap()                 // 42
err.UnwrapOr(0)             // 0

// From standard (value, error) pattern
result := g.ResultOf(strconv.Atoi("42"))    // Ok(42)
result := g.String("42").TryInt()           // Ok(42)

// Chaining
g.Ok(10).Then(func(x int) g.Result[int] {
    return g.Ok(x * 2)
}) // Ok(20)

// Convert to Option (discards error)
opt := result.Option()      // Some(42)
```

<details>
<summary><b>All Result Methods</b></summary>

| Method | Description |
|--------|-------------|
| `IsOk()` / `IsErr()` | State checks |
| `IsOkAnd(pred)` / `IsErrAnd(pred)` | State combined with a predicate |
| `Ok()` / `Err()` | Returns value / error (zero if the other state) |
| `Unwrap()` / `Expect(msg)` | Returns value, panics if Err |
| `UnwrapErr()` | Returns error, panics if Ok |
| `UnwrapOr(default)` / `UnwrapOrDefault()` | Returns value or fallback |
| `Map(fn)` | Transforms the value, may change its type |
| `MapOr(def, fn)` / `MapOrElse(defFn, fn)` | Map or fall back |
| `Then(fn)` / `ThenOf(fn)` | Chains Result / (T, error) functions |
| `Or(other)` / `OrElse(fn)` | Fallback Result |
| `MapErr(fn)` / `Wrap(err)` | Transforms / wraps the error |
| `Inspect(fn)` / `InspectErr(fn)` | Side effects, return self |
| `ErrIs(target)` / `ErrAs(target)` | errors.Is / errors.As integration |
| `Option()` / `Result()` | Converts to Option / (value, error) |

JSON: `Ok(v)` ⇄ `{"ok": v}`, `Err(e)` ⇄ `{"err": "msg"}` (externally tagged; duplicate keys rejected).

</details>

---

## Slice

Extended slice with 90+ methods.

```go
s := g.SliceOf(1, 2, 3, 4, 5)
s := g.NewSlice[int](0, 10)     // empty, with capacity 10 (one arg sets len AND cap)

s.Len()                         // Int(5)
s.Get(0)                        // Some(1)
s.Get(100)                      // None (safe!)
s.Last()                        // Some(5)
s.Contains(3)                   // true

s.Push(6, 7)                    // append in place
s.Pop()                         // Some(7)
s.Clone()                       // safe copy

// Sorting
s.SortBy(cmp.Cmp)               // ascending
s.Reverse()                     // in place
rand.Shuffle(s)                 // random order (g/rand)
```

<details>
<summary><b>All Slice Methods</b></summary>

| Category | Methods |
|----------|---------|
| **Access** | `Get`, `Last`, `First` |
| **Modify** | `Set`, `Push`, `Pop`, `Insert`, `Remove` |
| **Search** | `Contains`, `ContainsBy`, `Index`, `IndexBy` |
| **Transform** | `Clone`, `Reverse`, `SortBy`, `SubSlice` |
| **Convert** | `Iter`, `Std` |
| **Info** | `Len`, `Cap`, `IsEmpty` |

</details>

---

## Map

Extended map with Entry API.

```go
m := g.NewMap[string, int]()
// or from pairs: g.MapOf(g.Pair[string, int]{"a", 1}, g.Pair[string, int]{"b", 2})

m.Insert("a", 1)                // insert value
m.Get("a")                      // Some(1)
m.Get("x")                      // None
m.Contains("a")                 // true
m.Remove("a")                   // remove

m.Keys()                        // Slice of keys
m.Values()                      // Slice of values
```

### Entry API

Efficient update without multiple lookups:

```go
// Insert if absent
m.Entry("counter").OrInsert(0)

// Update if present, insert if absent
m.Entry("counter").AndModify(func(v *int) { *v++ }).OrInsert(1)

// Lazy initialization
m.Entry("key").OrInsertWith(func() int {
    return expensiveComputation()
})
```

**Word frequency example:**
```go
words := g.SliceOf("apple", "banana", "apple", "cherry", "banana", "apple")
freq := g.NewMap[string, int]()

for _, word := range words {
    freq.Entry(word).AndModify(func(v *int) { *v++ }).OrInsert(1)
}
// {"apple": 3, "banana": 2, "cherry": 1}
```

<details>
<summary><b>Entry Pattern Matching</b></summary>

```go
switch e := m.Entry("key").(type) {
case g.OccupiedEntry[string, int]:
    fmt.Println("Exists:", e.Get())
    e.Insert(newValue)      // replace
    e.Remove()              // delete
case g.VacantEntry[string, int]:
    e.Insert(defaultValue)  // insert
}
```

</details>

---

## Set

Collection of unique elements.

```go
s := g.SetOf(1, 2, 3)
s.Insert(4)                     // add
s.Remove(1)                     // remove
s.Contains(2)                   // true

// Set algebra (eager, returns Set)
s.Union(other)                  // A ∪ B
s.Intersection(other)           // A ∩ B
s.Difference(other)             // A \ B
s.SymmetricDifference(other)    // A △ B
s.Disjoint(other)               // no common elements?
```

### Set Operations

```go
a := g.SetOf(1, 2, 3, 4)
b := g.SetOf(3, 4, 5, 6)

a.Union(b)                      // Set{1, 2, 3, 4, 5, 6}
a.Intersection(b)               // Set{3, 4}
a.Difference(b)                 // Set{1, 2}
a.SymmetricDifference(b)        // Set{1, 2, 5, 6}

a.Subset(b)                     // false
a.Superset(b)                   // false
```

---

## Heap

Priority queue (binary heap).

```go
import "github.com/enetx/g/cmp"

// Min-heap (smallest first)
h := g.NewHeap(cmp.Cmp[int])
h.Push(5, 3, 8, 1, 9)
// or in one go: g.HeapOf(cmp.Cmp[int], 5, 3, 8, 1, 9)

h.Peek()                        // Some(1)
h.Pop()                         // Some(1)
h.Pop()                         // Some(3)

// Max-heap (largest first)
maxH := g.NewHeap(func(a, b int) cmp.Ordering {
    return cmp.Cmp(b, a)
})

// From a slice — spread into the variadic constructor
s := g.SliceOf(5, 3, 8, 1)
heap := g.HeapOf(cmp.Cmp[int], s...)
```

---

## Deque

Double-ended queue (ring buffer). O(1) at both ends.

```go
dq := g.NewDeque[int]()
dq := g.DequeOf(1, 2, 3)

dq.PushFront(0)                 // add to front
dq.PushBack(4)                  // add to back

dq.Front()                      // Some(0)
dq.Back()                       // Some(4)

dq.PopFront()                   // Some(0)
dq.PopBack()                    // Some(4)

dq.Get(1)                       // element at index
dq.Len()                        // length
dq.IsEmpty()                    // true/false
```

---

## Iterators

Functional operations on sequences.

```go
g.SliceOf(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).
    Iter().
    Filter(func(x int) bool { return x%2 == 0 }).
    Map(func(x int) int { return x * x }).
    Take(3).
    Collect().
    Slice()
// Slice[4, 16, 36]
```

Thanks to generic methods (Go 1.27+), transformations can change the element type
right in the chain:

```go
g.SliceOf(1, 2, 3).
    Iter().
    Map(func(x int) g.String { return g.Int(x).String() }). // Seq[int] -> Seq[g.String]
    Collect().
    Slice().
    Join(", ").
    Println() // 1, 2, 3

// Fold into any accumulator type
words := g.SliceOf[g.String]("a", "bb", "ccc")
total := words.Iter().Fold(g.Int(0), func(acc g.Int, w g.String) g.Int { return acc + w.Len() })
// 6

// Typed Zip — pairs keep their real types
nums := g.SliceOf(1, 2, 3)
names := g.SliceOf("one", "two", "three")
nums.Iter().Zip(names.Iter()).ForEach(func(n int, s string) {
    fmt.Println(n, s)
})
```

### Collecting

`Collect()` returns a collector; pick the container to materialize into. `Set`,
`Map`, `MapOrd` and `MapSafe` take explicit type arguments — a method of
`Seq[V any]` cannot constrain V to `comparable`, so the constraint lives on the
collector method and the type is spelled at the call site:

```go
s.Iter().Filter(fn).Collect().Slice()          // Slice[int]
s.Iter().Filter(fn).Collect().Set[int]()       // Set[int]
s.Iter().Collect().Deque()                     // *Deque[int]
s.Iter().Collect().Heap(cmp.Cmp)               // *Heap[int]

m.Iter().FilterByKey(fn).Collect().Pairs()               // []Pair[K, V]
m.Iter().FilterByKey(fn).Collect().Map[string, int]()    // Map[string, int]
m.Iter().Collect().MapOrd[string, int]()                 // MapOrd[string, int]
m.Iter().Collect().MapSafe[string, int]()                // *MapSafe[string, int]
```

Note: `Collect()` alone is lazy — nothing runs until a materializer is called.

### Common Patterns

```go
// Chain iterators
g.SliceOf(1, 2).Iter().Chain(g.SliceOf(3, 4).Iter()).Collect().Slice()
// Slice[1, 2, 3, 4]

// Sum with Fold
g.SliceOf(1, 2, 3, 4, 5).Iter().Fold(0, func(acc, x int) int { return acc + x })
// 15

// Enumerate
g.SliceOf("a", "b", "c").Iter().Enumerate().ForEach(func(i g.Int, v string) {
    fmt.Printf("%d: %s\n", i, v)
})

// Using f package predicates
import "github.com/enetx/g/f"

g.SliceOf(1, 2, 3, 4, 5).Iter().Filter(f.Ne(3)).Collect().Slice() // Slice[1, 2, 4, 5]
g.SliceOf("", "a", "").Iter().Exclude(f.IsZero).Collect().Slice()  // Slice["a"]
```

<details>
<summary><b>All Iterator Methods</b></summary>

| Transform | Slice | Combine | Aggregate | Search | Other |
|-----------|-------|---------|-----------|--------|-------|
| `Map` | `Take` | `Chain` | `Collect` | `Find` | `ForEach` |
| `Filter` | `Skip` | `Zip` | `Fold` | `Any` | `Inspect` |
| `Exclude` | `StepBy` | `Enumerate` | `Reduce` | `All` | `Range` |
| `FilterMap` | `First` | `Intersperse` | `Count` | `MaxBy` | `Scan` |
| `FlatMap` | `Last` | `Cycle` | `CounterBy` | `MinBy` | `Combinations` |
| `Flatten` | `Nth` | | `Partition` | | `Permutations` |
| `Dedup` | `Chunks` | | `ChunkBy` | | `Context` |
| `Unique` | `Windows` | | | | `Chan` |
| `SortBy` | `TakeWhile` | | | | |
| | `SkipWhile` | | | | |

</details>

---

## Mutex

Typed mutex — data bound to lock.

```go
counter := g.NewMutex(0)

// Lock and modify
guard := counter.Lock()
guard.Set(guard.Get() + 1)
guard.Unlock()

// With defer (recommended)
func increment(c *g.Mutex[int]) {
    guard := c.Lock()
    defer guard.Unlock()
    guard.Set(guard.Get() + 1)
}

// Direct pointer access
guard := counter.Lock()
defer guard.Unlock()
*guard.Deref() += 1

// Non-blocking
if opt := counter.TryLock(); opt.IsSome() {
    guard := opt.Unwrap()
    defer guard.Unlock()
    // got the lock
}
```

**Why typed mutex?**
```go
// Traditional — easy to forget locking
type Old struct {
    mu   sync.Mutex
    data map[string]int  // what protects this?
}

// Typed — impossible to access without lock
type New struct {
    data *g.Mutex[g.Map[string, int]]
}

func (s *New) Get(key string) g.Option[int] {
    guard := s.data.Lock()
    defer guard.Unlock()
    return guard.Deref().Get(key)
}
```

---

## RwLock

Multiple readers OR single writer.

```go
config := g.NewRwLock(Config{Port: 8080})

// Read (concurrent)
func getPort(c *g.RwLock[Config]) int {
    guard := c.Read()
    defer guard.Unlock()
    return guard.Get().Port
}

// Write (exclusive)
func setPort(c *g.RwLock[Config], port int) {
    guard := c.Write()
    defer guard.Unlock()
    guard.Deref().Port = port
}

// Non-blocking
if opt := config.TryRead(); opt.IsSome() { ... }
if opt := config.TryWrite(); opt.IsSome() { ... }
```

| Use Case | Choose |
|----------|--------|
| Read-heavy (config, cache) | `RwLock` |
| Write-heavy or balanced | `Mutex` |
| Simple counters | `Mutex` |

---

## Pool

Goroutine pool for parallel tasks.

```go
import "github.com/enetx/g/pool"

p := pool.New[int]().Limit(4)

for i := range 10 {
    p.Go(func() g.Result[int] {
        return g.Ok(i * 2)
    })
}

for result := range p.Wait() {
    if result.IsOk() {
        fmt.Println(result.Unwrap())
    }
}
```

<details>
<summary><b>With Context and Cancel on Error</b></summary>

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

p := pool.New[int]().
    Context(ctx).
    CancelOnError().
    Limit(4)

for i := range 100 {
    p.Go(func() g.Result[int] {
        if i == 50 {
            return g.Err[int](errors.New("error"))
        }
        return g.Ok(i)
    })
}

for result := range p.Wait() {
    // stops after first error
}
```

</details>

---

## String

Extended string with 80+ methods.

```go
s := g.String("Hello, World!")

s.Len()                         // 13
s.Upper()                       // "HELLO, WORLD!"
s.Lower()                       // "hello, world!"
s.Contains("World")             // true
s.Trim()                        // remove whitespace

// Splitters return a plain []String — index, len and range work directly
parts := s.Split(", ")          // []String{"Hello", "World!"}
parts[0]                        // "Hello"

for _, part := range s.Split(", ") {
    part.Println()              // Hello / World!
}

// For chains, lift the slice into an iterator
g.SliceOf(s.Split(", ")...).Iter().Map(g.String.Upper).Collect().Slice()

// Conversions
s.Std()                         // Go string
s.Bytes()                       // Bytes type
g.String("42").TryInt()         // Result[Int]

// Encoding
s.Hash().MD5()                  // hash
s.Encode().Base64()             // encode
s.Compress().Gzip()             // compress
```

### Randomness — `g/rand`

All randomness lives in one package (math/rand/v2 under the hood; the
`Secure*` helpers draw from crypto/rand):

```go
import "github.com/enetx/g/rand"

rand.N(10)                    // [0, 10)
rand.Range(5, 10)             // half-open [5, 10)
rand.RangeInclusive(1, 6)     // closed [1, 6]
rand.Float()                  // [0, 1)
rand.Chance(0.25)             // true with probability 0.25
rand.String(10)               // 10 alphanumeric characters
rand.SecureString(32)         // crypto/rand-backed token (keys, codes)
rand.SecureBytes(16)          // crypto/rand-backed raw bytes
rand.String(6, g.Digits)      // 6-digit code

users := g.SliceOf("ann", "bob", "eve")
rand.Choice(users)            // Option[string]
rand.Sample(users, 2)         // 2 distinct elements
rand.Shuffle(users)           // in place; accepts Slice, MapOrd, plain slices
```

The container types deliberately carry no random methods — one way to do it.

### Regular expressions — `g/rx`

Compile a pattern once, then run it anywhere — patterns live in their own package:

```go
import "github.com/enetx/g/rx"

re := rx.New(`[0-9]{4}-[0-9]{2}-[0-9]{2}`).Unwrap()

re.IsMatch("2026-08-09")                   // true
re.Find("on 2026-08-09 at noon")           // Option[String]
re.FindIter(hay).Take(2).Collect().Slice() // first two matches
re.Replace(s, "$1")                        // FIRST match only
re.ReplaceBy(s, fn)                        // FIRST match through a function
re.ReplaceAll(s, "$1")                     // every match

digits := rx.New(`[0-9]`).Unwrap()
digits.Split("a1b2c3").Collect().Slice()   // Slice[a, b, c]

// Bytes haystacks: IsMatchBytes, FindBytes, FindIterBytes, ...
```

---

## Int / Float

Numeric types with utility methods.

```go
i := g.Int(42)

i.Add(8)                        // 50
i.Mul(2)                        // 84
i.Abs()                         // absolute value
i.Min(10, 20)                   // minimum
i.Max(10, 20)                   // maximum

i.IsZero()                      // false
i.IsPositive()                  // true
i.IsNegative()                  // false

i.Binary()                      // binary string
i.Hex()                         // hex string

// Checked arithmetic — overflow becomes None instead of a silent wraparound
g.Int(math.MaxInt).CheckedAdd(1)   // None
g.Int(10).CheckedDiv(0)            // None — no division panic
g.Int(2).CheckedPow(62)            // Some(4611686018427387904)

// Saturating / Overflowing variants
g.Int(math.MaxInt).SaturatingAdd(1)   // MaxInt (clamped)
v, overflowed := g.Int(math.MaxInt).OverflowingAdd(1) // MinInt, true

// Float
f := g.Float(3.14159)
f.Round()                       // Int(3)
f.RoundDecimal(2)               // Float(3.14)
f.Sqrt()                        // square root

// Float math & classification
g.Float(1).Div(0).IsInf()       // true
g.Float(-3.75).Fract()          // -0.75
g.Float(2.5).Clamp(0, 2)        // 2
g.Float(math.Pi).ToDegrees()    // 180
g.Float(0.1).MulAdd(10, -1)     // true FMA — single rounding
// plus Exp/Ln/Log2/Log10, Sin..Atanh, Signum, Copysign, Hypot, Cbrt...

// Compare
i.Cmp(g.Int(50))                // cmp.Less
```

---

## Bytes

Extended byte slice.

```go
b := g.Bytes([]byte("Hello"))

b.Len()
b.Upper()
b.Lower()
b.Contains([]byte("ell"))

b.String()                      // to String
b.Std()                         // to []byte
```

---

## File / Directory

Filesystem types live in the `g/fs` subpackage, so only programs that touch
files pay for them:

```go
import "github.com/enetx/g/fs"
```

### File

```go
// Read/Write
content := fs.NewFile("config.json").Read()      // Result[String]
fs.NewFile("output.txt").Write("Hello")
fs.NewFile("log.txt").Append("line\n")

// Open a new file with the specified name "text.txt" and process it line by line.
fs.NewFile("text.txt").
	Lines().                            // Reads the file line by line.
	Skip(2).                            // Skips the first 2 lines in the iterator.
	Exclude(f.IsZero).                  // Excludes lines that are empty or contain only whitespaces.
	Dedup().                            // Removes consecutive duplicate lines.
	Map(g.String.Upper).                // Converts each line to uppercase.
	Range(func(s g.Result[g.String]) bool { // Iterates over the lines while a condition is true.
		if s.IsErr() { // Handles any errors encountered while reading lines.
			fmt.Println("Error:", s.Err())
			return false // Stops the iteration if an error occurs.
		}

		if s.Ok().Contains("COULD") { // Checks if the line contains the substring "COULD".
			return false // Stops the iteration if the condition is met.
		}

		fmt.Println(s.Ok()) // Prints the line.
		return true         // Continues the iteration.
	})

// Properties
file := fs.NewFile("test.txt")
file.Exists()                   // check existence
file.Stat()                     // file info
file.Copy("backup.txt")         // copy
file.Rename("new.txt")          // rename
file.Remove()                   // delete
```

### File Guard (Exclusive Lock)

```go
locked := fs.NewFile("data.txt").Guard()  // holds exclusive lock
locked.Write("exclusive data")
locked.Close()                            // release lock
```

### Directory

```go
fs.NewDir("mydir").Create()
fs.NewDir("path/to/deep").CreateAll()

// Read contents
for file := range fs.NewDir(".").Read() {
    if file.IsOk() {
        file.Ok().Name().Println()
    }
}

// Recursively walk through the directory tree starting from the current directory
fs.NewDir(".").Walk().
	// Exclude directories and symlinked directories
	Exclude(func(f *fs.File) bool { return f.IsDir() && f.Dir().Ok().IsLink() }).
	// Exclude file symlinks
	Exclude((*fs.File).IsLink).
	// Process each walk result
	ForEach(func(v g.Result[*fs.File]) {
		if v.IsOk() {
			// Print the path of the file if no error occurred
			v.Ok().Path().Ok().Println()
		}
	})

// Iterate over and print the names of files in the current directory with a *.go extension
fs.NewDir("*.go").Glob().ForEach(func(f g.Result[*fs.File]) { f.Ok().Name().Println() })

// Copy the contents of the current directory to a new directory named "copy".
fs.NewDir(".").Copy("copy").Unwrap()
```

---

## Formatting and Print

Placeholders support automatic, positional, and named values:

```go
g.Format("{} {2} {name}", "auto", "positional", g.Named{"name": "named"})
g.Format("{{{name}}}", g.Named{"name": "value"}) // {value}
```

Dotted placeholders access DATA — struct fields, map keys, slice indices —
never methods (transform values before formatting):

```go
g.Format("{1.Field}", structVal)          // struct field
g.Format("{.outer.inner}", nestedMap)     // map keys
g.Format("{.1.2}", nestedSlice)           // slice indices
```

Format specs include alignment, width, precision, signs, alternate prefixes,
and common representations:

```go
g.Format("{:d} {:c} {:q} {:U} {:#010x}", 42, 65, "go\n", 'A', 255)
```

Use `FormatTo` to append to a reusable builder. `TryFormat` validates templates
and returns `Result[String]`; `TryFormatTo` appends only when formatting succeeds.

Types can implement custom specs without reflection:

```go
type Price float64

func (p Price) FormatValue(spec g.String) g.String {
    if spec == "currency" {
        return g.String(fmt.Sprintf("$%.2f", p))
    }
    return g.String(fmt.Sprintf("%.2f", p))
}

g.Format("{:currency}", Price(19.95)) // $19.95
```

## Other Maps

### MapOrd — Ordered Map

Maintains insertion order.

```go
m := g.NewMapOrd[string, int]()
m.Insert("c", 3)
m.Insert("a", 1)
m.Insert("b", 2)

for k, v := range m.Iter() {
    fmt.Println(k, v)  // c, a, b order
}

m.SortByKey(cmp.Cmp)            // sort by key
m.SortByValue(cmp.Cmp)          // sort by value
```

### MapSafe — Concurrent Map

Thread-safe for concurrent access.

```go
m := g.NewMapSafe[string, int]()

// Safe from multiple goroutines
go func() { m.Insert("a", 1) }()
go func() { m.Get("a") }()
```

---

## License

MIT License
