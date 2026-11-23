# Contributing to go-brrr

Thank you for your interest in go-brrr. This document explains how to submit a
change that the maintainer can review, benchmark, and merge.

go-brrr is a performance-critical library. Most of the rules below exist to make
a performance claim reproducible.

## Before you start

Install the toolchain and the C reference library:

```sh
make init
```

`make init` pulls the `brotli-ref` submodule, installs `benchstat` and
`embedmd`, and builds `lib/libbrotli_cref.a`.

For a large optimization or an architectural change, open an issue or a draft
pull request first. Agree on the scope before you write the implementation.
This step protects your time, not only the review time.

## Pull request scope

**One pull request is one logical change.**

Submit independent optimizations, refactorings, algorithm changes, and changes
to separate subsystems as separate pull requests.

For a performance change:

- Give each independently applicable optimization its own pull request.
- Explain the workload and code path that the change targets.
- If two changes depend on each other, submit them together. Explain the
  dependency in the description.

Small pull requests can be reviewed, benchmarked, reverted, and bisected
independently. A pull request with eight independent changes produces one
number. That number does not show which change wins and which change loses.

The maintainer can ask you to split a large pull request before detailed
review. The maintainer can decline to review or merge a pull request until its
scope is smaller.

Open the first pull request and wait for feedback before you write the next
one. This avoids wasted work.

## Performance changes

Explain the expected effect and its cause. State known trade-offs.

You do not need to submit benchmark results. The maintainer measures performance
claims on dedicated hardware. The review includes microbenchmarks and the full
file corpus.

### Local benchmarks

You can use the repository harness to check your change before submission:

```sh
# Compare the current worktree against main. Profiles: enc, dec, ench.
./scripts/bench-compare.sh enc dec
```

The script builds a before binary from `main` and an after binary from your
worktree. It interleaves the two processes, collects samples, and runs
`benchstat`. Investigate each result across relevant workloads and code
alignments.

On a Linux machine with `sudo`, wrap the run in a denoise script. It disables
boost, sets the performance governor, and pins the run to one physical core:

```sh
./scripts/denoise-amd.sh ./scripts/bench-compare.sh enc
# or, on Intel:
./scripts/denoise-intel.sh ./scripts/bench-compare.sh enc
```

To diff hardware counters for a hot loop, use:

```sh
./scripts/perf-compare.sh 'BenchmarkMatchLenAtLong'
```

Useful environment variables: `COUNT`, `BENCHTIME`, `QUALITIES`, `PAYLOADS`,
`BASE_BRANCH`, `FUNCALIGNS`.

Code alignment can shift benchmark results by several percent on amd64. Use
`FUNCALIGNS` to compare several alignments before you attribute a result to the
code change.

## Correctness

Performance never wins over correctness. Run these checks before you open a
pull request:

```sh
go test ./...                 # with the SIMD paths
go test -tags purego ./...    # with the generic paths
```

Both must pass. CI runs both.

The compatibility tests compare the encoder output against the C reference
implementation, byte for byte. They need `lib/libbrotli_cref.a`.

Test output must be clean. If a test expects an error, it must assert that
error.

### Differential fuzzing

The targets in `fuzz_cref_test.go` compare Go with Google's C reference:

- `FuzzEncodeCRef` checks encoder equivalence and cross-decoding, including
  chunked writes, flushes, and streaming reads.
- `FuzzDecodeCRef` compares acceptance and decoded bytes. It accounts for
  trailing bytes: C ignores them, while Go returns `ErrExcessiveInput`.
- `FuzzCompoundDictCRef` checks encoding and decoding with a raw compound
  dictionary.

Encoder equivalence checks use a single write on both sides. Chunked writes
and flushes are checked by decoding with both implementations, since write
boundaries can change valid compressed output.

At quality 10–11, small rounding differences between Go's `math.Log2` and
the C math library's `log2` can change the encoder's estimated costs and
command choices. These differences can affect multiple commands, so a small
byte or percentage allowance is not a correctness guarantee for arbitrary
inputs. Without a compound dictionary, quality 5–9 with windows above 16
can also choose different matches in Go and C.

For these settings, fuzzing requires exact decoded bytes from both decoders
but imposes no compressed-size limit. Other settings retain byte-for-byte
encoder comparisons. Compression-size limits in the fixed-corpus tests are
regression budgets for those workloads, not correctness rules for arbitrary
fuzz inputs.

Build the pinned Google reference implementation, then explicitly enable
mutation fuzzing (ordinary `go test` only replays the seed corpus):

```sh
git submodule update --init
make lib/libbrotli_cref.a
go test -run='^$' -fuzz='^FuzzDecodeCRef$' -fuzztime=5m -parallel=4 .
go test -tags=purego -run='^$' -fuzz='^FuzzEncodeCRef$' -fuzztime=5m -parallel=4 .
```

Use `-fuzz='^FuzzCompoundDictCRef$'` to run the dictionary target. Fuzzing
requires cgo and a C compiler, including with `-tags=purego`. Plaintext and
dictionary inputs are capped at 64 KiB, compressed inputs at 8 KiB; these
limits do not cap decoded output. Go saves failing inputs under
`testdata/fuzz/<target>/`. Minimize and commit them only when they add coverage
that is not already captured by a focused regression test or an intentional
seed. CI runs the seed corpus on every push.

The `Nightly` workflow runs mutation fuzzing every night. On a failure it
uploads the failing inputs as an artifact.

## Code style

- Match the style of the file that you edit.
- Keep comments short. Explain why, not what.
- Do not write a comment about how the code changed. That belongs in git.
- Do not reformat lines that your change does not touch.

## AI-assisted contributions

You can use an AI tool to write code. The rules do not change:

- You understand the code that you submit.
- You ran the tests yourself.
- You can answer questions about the design in review.

Generated volume is not a contribution. A pull request that the author cannot
explain will be closed.

## Review expectations

A performance pull request takes a long time to review. The maintainer
benchmarks every change against a file corpus on dedicated hardware. This work
is serial and slow.

A small, focused pull request gets reviewed faster than a large one.

## What blocks a merge

The maintainer will request changes in these cases:

- The pull request contains several independent changes.
- The description does not explain the reason for the change.
- The `purego` build or the C reference compatibility tests fail.

The maintainer can close the pull request if the author does not address the
request.

## License

go-brrr uses the MIT license. Your contribution goes under the same license.
