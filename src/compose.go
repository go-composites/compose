package Compose

import (
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

/*
Step is a single stage of a railway-oriented pipeline. It receives the payload
produced by the previous stage and returns a Result. A Step never returns a
bare nil: on success it returns a payload-bearing Result, on failure a Result
whose HasError() is true.
*/
type Step func(input interface{}) Result.Interface

/*
Pipe composes steps left-to-right into a single reusable Step.

The composed Step threads a value down the track: each step receives the
payload of the previous step's Result. The instant a step returns a Result with
HasError() == true, the pipeline short-circuits and that error Result is
returned unchanged — no later step runs. Otherwise the final step's Result is
returned.

Pipe with no steps is the identity track: it wraps its input in a successful
Result. The returned Step is reusable and safe to run on many inputs.
*/
func Pipe(steps ...Step) Step {
	return func(input interface{}) Result.Interface {
		current := Result.New(Result.WithPayload(input))
		for _, step := range steps {
			current = step(current.Payload())
			if current.HasError() {
				return current
			}
		}
		return current
	}
}

/*
Run is a convenience for Pipe(steps...)(input): it composes the steps and runs
them on input in one call, returning the final (or short-circuited) Result.
*/
func Run(input interface{}, steps ...Step) Result.Interface {
	return Pipe(steps...)(input)
}

/*
Then lifts an ordinary, infallible transform into a Step. The transform runs on
the incoming payload and its return value becomes the payload of a successful
Result. Use it for the "happy path" stages that cannot fail.
*/
func Then(transform func(input interface{}) interface{}) Step {
	return func(input interface{}) Result.Interface {
		return Result.New(Result.WithPayload(transform(input)))
	}
}

/*
Map is an alias of Then, mirroring the functional-combinator vocabulary: it maps
a successful payload through transform.
*/
func Map(transform func(input interface{}) interface{}) Step {
	return Then(transform)
}

/*
Recover wraps a step so that, when it short-circuits with an error, handler is
given the chance to get the pipeline back onto the success track. handler runs
only on the error path; it receives the failed Result and returns a new Result
(e.g. a fallback payload, or the same error re-raised). On the success path the
step's Result passes through untouched.
*/
func Recover(step Step, handler func(failed Result.Interface) Result.Interface) Step {
	return func(input interface{}) Result.Interface {
		result := step(input)
		if result.HasError() {
			return handler(result)
		}
		return result
	}
}

/*
Fail builds a Step that always short-circuits with the given message. It is a
small helper for constructing failing stages without importing Error/Result at
the call site.
*/
func Fail(message string) Step {
	return func(input interface{}) Result.Interface {
		return Result.New(Result.WithError(Error.New(message)))
	}
}
