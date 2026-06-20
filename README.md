<p align="center"><img src="https://raw.githubusercontent.com/go-composites/brand/main/social/go-composites.png" alt="go-composites/compose" width="720"></p>

# Compose

`compose` is the centerpiece primitive of [go-composites](https://github.com/go-composites) — the operation that names the org. It provides **railway-oriented** pipelines: small steps composed left-to-right that thread a [`Result`](https://github.com/go-composites/result) along a success track and short-circuit onto an error track the instant something fails.

## Railway-oriented composition

Picture two parallel rails: a **success track** carrying a payload, and an **error track** carrying a failure. Every step is a switch:

- given a payload, a step either stays on the success track (produces a payload-bearing `Result`), or
- diverts onto the error track (produces a `Result` whose `HasError()` is `true`).

`Pipe` wires steps together. As long as steps stay on the success rail, each one's payload feeds the next. The moment a step diverts, the whole pipeline short-circuits: the remaining steps are skipped and the error `Result` is returned unchanged. This replaces nested `if err != nil` ladders with a single linear flow, and — true to the go-composites null-object discipline — **a step or pipe never returns a bare `nil`**; it always returns a real `Result`.

## Install

```sh
go get github.com/go-composites/compose@main
```

## API

```go
type Step func(input interface{}) Result.Interface

func Pipe(steps ...Step) Step                 // compose left-to-right; short-circuit on error
func Run(input interface{}, steps ...Step) Result.Interface  // Pipe(steps...)(input)

func Then(transform func(interface{}) interface{}) Step  // lift an infallible transform onto the success track
func Map(transform func(interface{}) interface{}) Step   // alias of Then
func Recover(step Step, handler func(Result.Interface) Result.Interface) Step  // handle the error track
func Fail(message string) Step                // a step that always diverts to the error track
```

Semantics:

- **`Pipe(steps...)`** returns a reusable `Step`. Running it wraps the input in a successful `Result`, then runs each step on the previous step's `Payload()`. If any step's `HasError()` is true, that error `Result` is returned immediately and later steps do not run. With no steps it is the identity track (wraps the input).
- **`Run(input, steps...)`** is `Pipe(steps...)(input)`.
- **`Then` / `Map`** lift a plain transform `func(interface{}) interface{}` into a `Step` that cannot fail.
- **`Recover(step, handler)`** runs `handler` only when `step` diverts to the error track, giving you a chance to return a fallback `Result` and get back on the success track. On success the `Result` passes through untouched.
- **`Fail(message)`** is a `Step` that always diverts with the given message.

## Usage

```go
import (
	Compose "github.com/go-composites/compose/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

func parse(input interface{}) Result.Interface {
	n, err := strconv.Atoi(input.(string))
	if err != nil {
		return Result.New(Result.WithError(Error.New("not an integer")))
	}
	return Result.New(Result.WithPayload(n))
}

func validate(input interface{}) Result.Interface {
	if input.(int) <= 0 {
		return Result.New(Result.WithError(Error.New("must be positive")))
	}
	return Result.New(Result.WithPayload(input))
}

// A reusable pipeline: parse -> validate -> double.
pipeline := Compose.Pipe(
	parse,
	validate,
	Compose.Then(func(input interface{}) interface{} { return input.(int) * 2 }),
)

ok := pipeline("21")            // HasError=false, Payload=42
bad := pipeline("not-a-number") // HasError=true,  Error().Message()="not an integer" (validate & double skipped)
```

See [`main.go`](main.go) for a runnable demo of both the success path and the short-circuit-on-error path.

## License

BSD-3-Clause. See [LICENSE](LICENSE).
