// Package g provides an ergonomic standard library extension for Go:
// monadic error handling, rich generic containers, lazy iterators with
// type-changing generic methods (Go 1.27+), and ergonomic wrappers for
// primitive types and the filesystem.
//
// # Monads
//
// [Option] represents an optional value (Some/None) and [Result] represents
// success or failure (Ok/Err). Both offer chainable combinators — Map, Then,
// UnwrapOr, MapOr, Inspect and friends — so error and nil handling become
// expressions instead of if-ladders.
//
// # Containers
//
//   - [Slice]: an extended slice with 90+ methods.
//   - [Map]: a map with ergonomic accessors and entry API.
//   - [MapOrd]: an insertion-ordered map.
//   - [MapSafe]: a concurrency-safe map.
//   - [Set]: a hash set with algebraic operations (union, intersection, ...).
//   - [Deque]: a double-ended queue backed by a ring buffer.
//   - [Heap]: a binary min/max heap driven by a comparison function.
//
// # Iterators
//
// Every container exposes an Iter method returning lazy sequences (Heap
// additionally offers a draining IntoIter) —
// [Seq] for values, [Seq2] for key-value pairs, [SeqSlices] for grouped
// windows/chunks and [SeqResult] for fallible streams — built on g's own
// dependency-free iterator core. With Go 1.27 generic methods, transformations
// can change the element type mid-chain:
//
//	g.SliceOf(1, 2, 3).Iter().Map[string](strconv.Itoa).Collect().Slice() // Slice[string]
//
// Iterators are lazy: nothing is computed until a consumer (Collect, ForEach,
// Fold, ...) runs the chain.
//
// # Primitive wrappers
//
// [String], [Int], [Float] and [Bytes] wrap the built-in types with fluent
// methods, including conversion pipelines via Encode/Decode (Base64, Hex,
// Octal, Binary, JSON, ...), Compress/Decompress (gzip, zlib, flate)
// and Hash (MD5, SHA1, SHA256, SHA512).
//
// # Subpackages
//
//   - fs: File and Dir — chainable, Result-based file and directory
//     operations, including lazy line/chunk iterators over file contents.
//   - rx: compiled regular expressions with a compile-once, rich matching API.
//   - rand: all randomness (numbers, choices, samples, shuffles, strings)
//     in one place.
//   - pool: a generic goroutine pool with limits, rate limiting and streaming.
//   - cmp: ordering primitives (cmp.Ordering, cmp.Cmp) used by sorts and heaps.
//   - f: predicate combinators for filters (f.Eq, f.Gt, f.Contains, ...).
//   - ref: pointer helpers (ref.Of).
//   - constraints: generic type constraints shared across the library.
//   - dbg: debugging helpers that print expressions with source locations.
//
// # Panics
//
// The library reports failures through [Result] and [Option]; panics are
// reserved for programmer errors. Only two families of API panic:
//
//   - the Unwrap/Expect family on [Option] and [Result] (Unwrap, UnwrapErr,
//     Expect) when called on the wrong variant;
//   - documented index-based operations (e.g. Slice.Swap/Insert/Replace/SubSlice,
//     Deque.Insert/Swap) when given out-of-range indices, and constructors with
//     documented preconditions (e.g. NewHeap with a nil comparison function).
//
// Everything else returns Option/Result instead of panicking.
//
// # Security
//
// [Format] placeholders resolve data only: each dot-segment selects a map
// key, a MapOrd key, a slice/array index or a struct field — placeholders
// cannot invoke methods. Note that an untrusted template string can still
// choose which of the supplied argument data is printed.
package g
