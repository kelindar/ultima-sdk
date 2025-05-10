# Go Coding Standards

This document outlines the coding standards and best practices for Go projects. It's structured into, each focusing on a different aspect of Go development. The rules are designed to ensure code quality, maintainability, and performance.

# 1 Data Orientation & Memory Layout

1. **MUST** store homogeneous data contiguously to maximise cache‑line hits.
   Example:

   ```go
   // Structure‑of‑Arrays (SoA) for scans
   type column struct{ values []float64 }
   ```

2. **MUST** pre‑allocate slices when the final length is known.
   Example:

   ```go
   buf := make([]byte, 0, 8*1024) // avoids realloc
   ```

3. **MUST** keep structs ≤ 64 B when passed on hot paths.
   Example:

   ```go
   type cell struct{ x, y uint32; id uint32 } // 12 B
   ```

4. **MUST** pad structs that live in different goroutines to one cache line.
   Example:

   ```go
   type counter struct{ v int64; _ [56]byte } // 64 B
   ```

5. **MUST** batch copies with the built‑in `copy` for memcpy‑speed moves.
   Example:

   ```go
   copy(dst, src[:n])
   ```

6. **NEVER** allocate pointer‑rich graphs on the hot path.
   Example:

   ```go
   // ✗ avoid on critical path
   type node struct { next *node; payload []byte }
   ```

7. **MUST** favour SoA over AoS for analytical loops.
   Example:

   ```go
   distances := coords.x.Sub(other.x) // vectorised diff
   ```

8. **MUST** reuse buffers via pooling only after profiling shows GC pressure > 5 %.
   Example:

   ```go
   buf := pool.Get().([]byte)
   ```

9. **MUST** align fixed arrays to word boundaries when using atomics.
   Example:

   ```go
   var flags [256]uint64 // naturally 8‑byte aligned
   ```

10. **NEVER** cast between incompatible struct layouts with `unsafe`.
    Example:

    ```go
    // ✗ undefined behaviour
    hdr := (*reflect.StringHeader)(unsafe.Pointer(&myStruct))
    ```

---

# 2 API Design

1. **MUST** make the zero value of every type immediately usable.
   Example:

   ```go
   var m intmap.Map; v, ok := m.Get(2)
   ```

2. **MUST** accept interfaces and return concrete types (except `error`).
   Example:

   ```go
   func Encode(w io.Writer, v any) error
   ```

3. **MUST** use functional options for optional configuration.
   Example:

   ```go
   db := column.New(column.WithCapacity(1<<20))
   ```

4. **NEVER** expose `interface{}` in a public signature.
   Example:

   ```go
   // ✗ do not
   func Process(data interface{}) {}
   ```

5. **MUST** return `(T, error)`—error last, no dual success flags.
   Example:

   ```go
   func Parse(b []byte) (Block, error)
   ```

6. **MUST** add `context.Context` to any operation that can block.
   Example:

   ```go
   func (c *Client) Fetch(ctx context.Context, id string) (Item, error)
   ```

7. **MUST** use pointer receivers for mutating methods; value receivers for pure reads.
   Example:

   ```go
   func (m *Map) Set(k, v uint32)
   func (m Map) Len() int
   ```

8. **MUST** provide explicit `Clone()` for deep copies and document cost.
   Example:

   ```go
   func (b Bitmap) Clone() Bitmap // O(n)
   ```

9. **NEVER** leak internal fields; guard invariants with unexported members.
   Example:

   ```go
   type Index struct{ items []entry } // unexported slice
   ```

10. **SHOULD** suffix read‑only accessors with nothing and mutators with a verb.
    Example:

    ```go
    size := b.Size()    // getter
    b.SetSize(1024)     // setter
    ```

---

# 3 Naming & Declaration

1. **MUST** use a single lowercase noun for package names.
   Example:

   ```go
   package bitmap
   ```

2. **MUST** CamelCase every exported identifier.
   Example:

   ```go
   type IntMap struct{}
   ```

3. **MUST** keep generic parameters to a single capital.
   Example:

   ```go
   func Max[T cmp.Ordered](a, b T) T
   ```

4. **NEVER** use Hungarian notation or type prefixes.
   Example:

   ```go
   // ✗ avoid
   var szLen int
   ```

5. **MUST** use one‑ or two‑letter names in tight loops.
   Example:

   ```go
   for i := 0; i < n; i++ {
       sum += xs[i]
   }
   ```

6. **SHOULD** avoid uncommon abbreviations; stick to Go community norms.
7. **MUST** prefix test helpers with `test`.
   Example:

   ```go
   func testClock() time.Time
   ```

8. **MUST** group related constants using `iota`.
   Example:

   ```go
   const (
       statusOK = iota
       statusErr
   )
   ```

9. **NEVER** create stutter names (`bitmap.BitmapBitmap`).
10. **SHOULD** annotate arch/purego files with `//go:build` tags.
    Example:

    ```go
    //go:build !purego && amd64
    ```

---

# 4 Control Flow & Readability

1. **MUST** return early so the happy path stays flush‑left.
   Example:

   ```go
   if err := validate(x); err != nil {
       return err
   }
   process(x)
   ```

2. **MUST** keep nesting depth ≤ 2; refactor deeper logic.
   Example:

   ```go
   if cond {
       if sub { return foo() }
   }
   ```

3. **MUST** tag branches with clear predicates.
   Example:

   ```go
   switch state {
   case open:
   case closed:
   }
   ```

4. **NEVER** rely on label `goto` except in low‑level FSMs.
5. **NEVER** use `for {}` with `break` just for scoping.
6. **MUST** separate variable declaration and large initialiser logic.
7. **SHOULD** name booleans positively (`done` not `notDone`).
8. **MUST** guard deferred unlocks only in non‑hot paths.
9. **MUST** avoid in‑place slice modifications inside range loops.
10. **SHOULD** move long anonymous funcs into named helpers for legibility.

---

# 5 Error Handling

1. **MUST** wrap errors with `%w` to preserve causal chain.
   Example:

   ```go
   return fmt.Errorf("open index: %w", err)
   ```

2. **MUST** use `errors.Is` / `errors.As` for checks.
3. **NEVER** `panic` on recoverable errors—reserve for programmer bugs.
4. **MUST** embed actionable context (“id 123 not found”).
5. **SHOULD** declare a named return `err` when mixing defer+error.
6. **MUST** choose between `(bool)` or `error`, not both.
   Example:

   ```go
   value, ok := cache.Get(k) // ok pattern
   ```

7. **SHOULD** convert `io.EOF` into package sentinel when it’s not fatal.
8. **NEVER** ignore errors (`_ = f.Close()`).
9. **MUST** document retryability via sentinel (`ErrTemporary`).
10. **SHOULD** unify error codes in a single file per package.

---

# 6 Concurrency

1. **MUST** shard data and lock per shard, not globally.
   Example:

   ```go
   idx := hash(k) & (shards - 1)
   mu[idx].Lock()
   ```

2. **MUST** tie goroutine lifetime to `context.Context`.

   ```go
   select { case <-ctx.Done(): return }
   ```

3. **NEVER** launch fire‑and‑forget goroutines in a library.
4. **MUST** use atomics for simple counters.

   ```go
   atomic.AddUint64(&c, 1)
   ```

5. **SHOULD** avoid busy‑spin select defaults.
6. **MUST** cap worker pools via buffered channels.
7. **NEVER** synchronise with `time.Sleep`.
8. **MUST** lock resources in a globally consistent order.
9. **SHOULD** expose `Close()` to drain goroutines.
10. **MUST** run `go test -race`—failures block merge.

---

# 7 Performance Optimisation

1. **MUST** benchmark changes; > 3 % regressions rejected.
   Example:

   ```bash
   benchstat old.txt new.txt
   ```

2. **MUST** inline tiny hot functions (`//go:inline`).
3. **SHOULD** hoist invariants outside loops.
4. **NEVER** allocate in tight loops; ensure 0 B/op.
5. **MUST** prefer constant‑time algorithms over logarithmic when dataset fits cache.
6. **MUST** collapse bounds checks via full‑capacity slicing (`buf[:n:n]`).
7. **SHOULD** use bulk `copy` for data moves.
8. **MUST** benchmark against stdlib baselines.
9. **NEVER** micro‑optimise without profiling evidence.
10. **SHOULD** supply arch‑specific SIMD/asm with generic fallback.

---

# 8 Comments & Documentation

1. **MUST** start comment with identifier name.
   Example:

   ```go
   // Encode writes v to w in binary format.
   ```

2. **MUST** use present tense, single‑sentence summary first.
3. **MUST** document complexity (`// O(n)` if > O(1)).
4. **MUST** add ASCII diagrams for non‑trivial algorithms.
5. **SHOULD** cite papers or RFCs.
6. **NEVER** comment what the code already states.
7. **MUST** update comments as part of every change.
8. **MUST** include package overview in `doc.go`.
9. **SHOULD** outline invariants and contracts.
10. **NEVER** leave stale TODOs.

---

# 9 Testing & Benchmarking

1. **MUST** write table‑driven tests for each exported API.
2. **MUST** keep algorithm packages ≥ 95 % coverage.
3. **MUST** enable `-race` in CI.
4. **MUST** keep benchmarks in `/bench`; commit golden stats.
5. **SHOULD** add property tests for `unsafe` code paths.
6. **NEVER** use sleeps for timing asserts.
7. **MUST** seed random with fixed value in tests.
8. **MUST** assert no allocs in critical benchmarks (`testing.AllocsPerRun`).
9. **SHOULD** stub external systems—tests must be hermetic.
10. **NEVER** ignore failing tests with `t.Skip()` on mainline.

---

# 10 Unsafe & Assembly

1. **MUST** confine `unsafe` to `/internal`.
2. **MUST** state memory invariants before every `unsafe` block.
   Example:

   ```go
   // ptr points to len(bytes) valid bytes, never reallocated during use.
   ```

3. **NEVER** traverse beyond allocation bounds via pointer math.
4. **MUST** offer pure‑Go fallback behind `//go:build purego`.
5. **SHOULD** prefer intrinsics before handwritten asm.
6. **MUST** pair asm files with one Go stub each.
7. **MUST** document endianness assumptions.
8. **NEVER** mix calling conventions in the same package.
9. **MUST** cover unsafe paths under `go test -race`.
10. **SHOULD** property‑test memory‑safety invariants.
