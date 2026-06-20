package main

import (
	"fmt"
	"strconv"
	"strings"

	Compose "github.com/go-composites/compose/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

// parse turns the incoming string into an integer, short-circuiting with an
// error Result when the text is not a number.
func parse(input interface{}) Result.Interface {
	text := strings.TrimSpace(input.(string))
	n, err := strconv.Atoi(text)
	if err != nil {
		return Result.New(Result.WithError(Error.New("parse: " + strconv.Quote(text) + " is not an integer")))
	}
	return Result.New(Result.WithPayload(n))
}

// validate rejects non-positive numbers, again as an error Result.
func validate(input interface{}) Result.Interface {
	n := input.(int)
	if n <= 0 {
		return Result.New(Result.WithError(Error.New("validate: value must be positive")))
	}
	return Result.New(Result.WithPayload(n))
}

// transform doubles the validated number. It cannot fail, so it is lifted with
// Compose.Then below rather than written as a raw Step.
func transform(input interface{}) interface{} {
	return input.(int) * 2
}

func main() {
	// A reusable pipeline: parse -> validate -> (double).
	pipeline := Compose.Pipe(
		parse,
		validate,
		Compose.Then(transform),
	)

	fmt.Println("== success path ==")
	ok := pipeline("21")
	fmt.Printf("input %q -> HasError=%t payload=%v\n", "21", ok.HasError(), ok.Payload())

	fmt.Println("== short-circuit: bad parse ==")
	badParse := pipeline("not-a-number")
	fmt.Printf("input %q -> HasError=%t error=%q\n",
		"not-a-number", badParse.HasError(), badParse.Error().Message())

	fmt.Println("== short-circuit: failed validation ==")
	badValidate := pipeline("-5")
	fmt.Printf("input %q -> HasError=%t error=%q\n",
		"-5", badValidate.HasError(), badValidate.Error().Message())

	fmt.Println("== Run convenience + Recover ==")
	recovered := Compose.Run("oops",
		Compose.Recover(parse, func(failed Result.Interface) Result.Interface {
			// fall back to a default value when parsing fails
			return Result.New(Result.WithPayload(0))
		}),
	)
	fmt.Printf("input %q -> HasError=%t payload=%v (recovered)\n",
		"oops", recovered.HasError(), recovered.Payload())
}
