package Compose_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	Compose "github.com/go-composites/compose/src"
	Error "github.com/go-composites/error/src"
	Result "github.com/go-composites/result/src"
)

// incr is a Step that adds one to an int payload, always succeeding.
func incr(input interface{}) Result.Interface {
	return Result.New(Result.WithPayload(input.(int) + 1))
}

// boom is a Step that always short-circuits with a sentinel error.
func boom(input interface{}) Result.Interface {
	return Result.New(Result.WithError(Error.New("boom")))
}

var _ = Describe("Compose", func() {
	Describe("Pipe", func() {
		It("threads the payload through every step on the success path", func() {
			out := Compose.Pipe(incr, incr, incr)(0)
			Expect(out).NotTo(BeNil())
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal(3))
		})

		It("is the identity track when given no steps", func() {
			out := Compose.Pipe()(7)
			Expect(out).NotTo(BeNil())
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal(7))
		})

		It("short-circuits and returns the error Result, skipping later steps", func() {
			ran := false
			after := func(input interface{}) Result.Interface {
				ran = true
				return Result.New(Result.WithPayload(input))
			}
			out := Compose.Pipe(incr, boom, after)(0)
			Expect(out).NotTo(BeNil())
			Expect(out.HasError()).To(BeTrue())
			Expect(out.Error().Message()).To(Equal("boom"))
			Expect(ran).To(BeFalse(), "the step after the failure must not run")
		})

		It("returns a reusable Step", func() {
			pipe := Compose.Pipe(incr, incr)
			Expect(pipe(0).Payload()).To(Equal(2))
			Expect(pipe(10).Payload()).To(Equal(12))
		})
	})

	Describe("Run", func() {
		It("composes and runs in one call on the success path", func() {
			out := Compose.Run(1, incr, incr)
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal(3))
		})

		It("propagates a short-circuit error", func() {
			out := Compose.Run(1, boom)
			Expect(out.HasError()).To(BeTrue())
			Expect(out.Error().Message()).To(Equal("boom"))
		})
	})

	Describe("Then / Map", func() {
		It("lifts an infallible transform into a successful Step (Then)", func() {
			out := Compose.Then(func(input interface{}) interface{} {
				return input.(int) * 10
			})(4)
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal(40))
		})

		It("Map behaves like Then", func() {
			out := Compose.Map(func(input interface{}) interface{} {
				return input.(string) + "!"
			})("hi")
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal("hi!"))
		})
	})

	Describe("Recover", func() {
		It("passes a successful Result through untouched", func() {
			handlerRan := false
			step := Compose.Recover(incr, func(failed Result.Interface) Result.Interface {
				handlerRan = true
				return failed
			})
			out := step(0)
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal(1))
			Expect(handlerRan).To(BeFalse())
		})

		It("runs the handler on the error path and can get back on track", func() {
			step := Compose.Recover(boom, func(failed Result.Interface) Result.Interface {
				Expect(failed.HasError()).To(BeTrue())
				return Result.New(Result.WithPayload("fallback"))
			})
			out := step(0)
			Expect(out.HasError()).To(BeFalse())
			Expect(out.Payload()).To(Equal("fallback"))
		})
	})

	Describe("Fail", func() {
		It("always short-circuits with the given message", func() {
			out := Compose.Fail("nope")(123)
			Expect(out).NotTo(BeNil())
			Expect(out.HasError()).To(BeTrue())
			Expect(out.Error().Message()).To(Equal("nope"))
		})
	})
})
