package retry_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/acorello/retry"
	"github.com/matryer/is"
)

func TestFn(t *testing.T) {
	type Given struct {
		retry.Spec
		Result string
		Errors []error
	}
	type Wanted struct {
		Calls int
	}

	var testCases = map[string]struct {
		Given
		Wanted
	}{
		"no error and default spec results in one invocation": {
			Given{
				Spec: retry.Spec{
					Retries: 0,
					Stop:    nil,
				},
				Result: "result",
				Errors: nil,
			},
			Wanted{
				Calls: 1,
			},
		},
		"constant error and no Stop results in max retries": {
			Given{
				Spec: retry.Spec{
					Retries: 3,
					Stop:    nil,
				},
				Result: "result",
				Errors: slices.Repeat([]error{fmt.Errorf("const error")}, 9),
			},
			Wanted{
				Calls: 1 + 3, // first attempt and 3 retries
			},
		},
		"one error of two allowed and no Stop results in one retry": {
			Given{
				Spec: retry.Spec{
					Retries: 2,
					Stop:    nil,
				},
				Result: "result",
				Errors: slices.Repeat([]error{fmt.Errorf("const error")}, 1),
			},
			Wanted{
				Calls: 1 + 1,
			},
		},
		"constant error with immediate stop results in no retries": {
			Given{
				Spec: retry.Spec{
					Retries: 4,
					Stop:    stopAfter(1),
				},
				Result: "result",
				Errors: slices.Repeat([]error{fmt.Errorf("const error")}, 9),
			},
			Wanted{
				Calls: 1,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			its := is.New(t)

			var retryCount int
			fnSpy := func() (string, error) {
				retryCount++
				return tc.Given.Result, SafeGet(tc.Errors, retryCount-1)
			}

			actualResult, actualErr := retry.Fn[string](fnSpy, tc.Given.Spec)

			its.Equal(tc.Given.Result, actualResult)
			its.Equal(SafeGet(tc.Errors, retryCount-1), actualErr)
			its.Equal(tc.Wanted.Calls, retryCount)
		})
	}
}

func SafeGet[T any](slice []T, index int) T {
	var zero T
	if index < 0 || index >= len(slice) {
		return zero
	}
	return slice[index]
}

func stopAfter(invocations uint) func(err error) bool {
	if invocations == 0 {
		panic("invocations count must be at least 1")
	}
	return func(err error) bool {
		invocations--
		return invocations == 0
	}
}
