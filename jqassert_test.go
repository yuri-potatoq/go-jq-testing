package jqassert_test

import (
	"errors"
	jqassert "github.com/yuri-potatoq/jq-assert"
	"testing"
)

func assertErr(t *testing.T, err, expectedErr error) {
	if !(errors.Is(err, expectedErr) || errors.Is(expectedErr, err)) {
		t.Errorf("expected error [%+v] not eq to given error [%+v]", expectedErr, err)
	}
}

func TestSimpleMatchArrays(t *testing.T) {
	targetJson := `{
	"main_contacts": [ "0000-0000" ],
	"house_numbers_unsorted": [ 201, 20, 3031 ],
	"house_numbers_sorted": [ 2020, 3030, 4040 ]
}`

	for _, tc := range []struct {
		description   string
		expr          string
		expectedErr   error
		assertionOpts []jqassert.AssertExpectation
	}{
		{
			description:   "SHOULD NOT raise ERROR for idx 0 of array",
			expr:          ".main_contacts[0]",
			assertionOpts: []jqassert.AssertExpectation{jqassert.ExistsAssertion},
		},
		{
			description:   "SHOULD raise ExistsAssertionErr for idx 1 of array",
			expr:          ".main_contacts[1]",
			assertionOpts: []jqassert.AssertExpectation{jqassert.ExistsAssertion},
			expectedErr:   jqassert.ExistsAssertionErr,
		},
		{
			description:   "SHOULD raise error for count of items equals 1",
			expr:          ".main_contacts[1]",
			assertionOpts: []jqassert.AssertExpectation{jqassert.TotalAssertion(1)},
			expectedErr:   jqassert.TotalAssertionErr,
		},
		{
			description:   "SHOULD raise TotalAssertionErr for count of items equals 2",
			expr:          ".main_contacts[1]",
			assertionOpts: []jqassert.AssertExpectation{jqassert.TotalAssertion(2)},
			expectedErr:   jqassert.TotalAssertionErr,
		},
		{
			description:   "SHOULD raise BooleanAssertionErr for unsorted array elements",
			expr:          ".house_numbers_unsorted | . == ( . | sort)",
			assertionOpts: []jqassert.AssertExpectation{jqassert.BooleanAssertion},
			expectedErr:   jqassert.BooleanAssertionErr,
		},
		{
			description:   "SHOULD NOT raise error for sorted array elements",
			expr:          ".house_numbers_sorted | . == ( . | sort)",
			assertionOpts: []jqassert.AssertExpectation{jqassert.BooleanAssertion},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			assertErr(t, jqassert.Assert(targetJson, tc.expr, tc.assertionOpts...), tc.expectedErr)
		})
	}

}
