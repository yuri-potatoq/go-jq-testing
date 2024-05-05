package jqassert

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itchyny/gojq"
	"testing"
)

type AssertionResultError struct {
	// Generic error returned for parsing or syntax issues
	assertionErr error

	// List of errors for expectations checks
	expectationsErr error
}

func (assR *AssertionResultError) Is(target error) bool {
	if assR == nil {
		return errors.Is(nil, target)
	}
	return errors.Is(assR.assertionErr, target) || errors.Is(assR.expectationsErr, target)
}

func (assR *AssertionResultError) Error() string {
	var assertionErrs = errors.Join(assR.expectationsErr, assR.assertionErr)

	//TODO: prettify the error message
	return fmt.Sprintf("Not all expections are correct: %s", assertionErrs)
}

// Return the check result and the check description
type AssertExpectation func(results []any) (bool, error)

func assert(plainJson, expr string, expecFs ...AssertExpectation) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return &AssertionResultError{
			assertionErr: fmt.Errorf("assertion fails with unexpeted query syntax: %s", err),
		}
	}

	var inp map[string]any
	err = json.Unmarshal([]byte(plainJson), &inp)
	if err != nil {
		return &AssertionResultError{
			assertionErr: fmt.Errorf("assertion fails with unexpeted json syntax: %s", err),
		}
	}

	iter := query.Run(inp)
	matchResults := make([]any, 0)

	resultErr := new(AssertionResultError)
	for {
		v, ok := iter.Next()
		if v == nil && ok == false {
			// execution has finished
			for _, expcf := range expecFs {
				if ok, expcErr := expcf(matchResults); !ok {
					resultErr.expectationsErr = errors.Join(
						resultErr.expectationsErr,
						//TODO: format results
						fmt.Errorf("MatchResults: [%+v]: %w", matchResults, expcErr))
				}
			}
			if resultErr.expectationsErr != nil {
				return resultErr
			}
			return nil
		}
		if v == nil && ok == true {
			// step finished with no matchResults
			continue
		}

		switch v.(type) {
		case error, *gojq.HaltError:
			return &AssertionResultError{
				assertionErr: fmt.Errorf("can't get the value: %s", err),
			}
		default:
			// has value with no error, append and continue
			matchResults = append(matchResults, v)
		}
	}
}

// Assertions functions errors
var (
	ExistsAssertionErr  = errors.New("exists/assertion: no values brought by the expression")
	BooleanAssertionErr = errors.New("boolean/assertion: ")
	TotalAssertionErr   = errors.New("total/assertion: brought amount of values doesn't match")
)

var (
	// True if the result is not null
	ExistsAssertion = func(results []any) (bool, error) {
		return len(results) > 0, ExistsAssertionErr
	}

	// Check by the boolean result of expression
	BooleanAssertion = func(results []any) (bool, error) {
		if len(results) != 1 {
			return false, fmt.Errorf("%w: the result should be a single boolean", BooleanAssertionErr)
		}

		result, ok := results[0].(bool)
		if !ok {
			return false, fmt.Errorf("%w: the result should by of boolean type", BooleanAssertionErr)
		}
		return result, fmt.Errorf("%w: the result boolean is false", BooleanAssertionErr)
	}

	// True if the total result items is equal to expected total
	TotalAssertion = func(total int) AssertExpectation {
		return func(results []any) (bool, error) {
			return len(results) == total, TotalAssertionErr
		}
	}
)

func Assert(plainJson, expr string, expecFs ...AssertExpectation) error {
	return assert(plainJson, expr, expecFs...)
}

func AssertExits(t *testing.T, plainJson, expr string) {
	triggerTestErr(t, assert(plainJson, expr, ExistsAssertion))
}

func AssertTotal(t *testing.T, plainJson, expr string, total int) {
	triggerTestErr(t, assert(plainJson, expr, TotalAssertion(total)))
}

func triggerTestErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
