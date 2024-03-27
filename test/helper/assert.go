//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2024 redd
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Directly taken from what I made for git.xdol.org/xdol/hashdog

package helper

import (
	"errors"
	"testing"
)

type Adapter struct{ t *testing.T }

func NewAdapter(t *testing.T) Adapter {
	t.Helper()

	return Adapter{t: t}
}

// Assert tests if two values are different, print an error if they are.
func (suite Adapter) Assert(got, expected any) {
	suite.t.Helper()
	if got != expected {
		suite.t.Errorf("Got: %v | Expected: %v\n", got, expected)
	}
}

// Assertf tests if two values are different, fail tests if they are.
func (suite Adapter) Assertf(got, expected any) {
	suite.t.Helper()
	if got != expected {
		suite.t.Fatalf("Got: %v | Expected: %v\n", got, expected)
	}
}

// AssertNotEmpty tests if the value corresponds to its empty value, print an error if it's the case.
func (suite Adapter) AssertNotEmpty(got any, emptyValue any) {
	suite.t.Helper()
	if got == emptyValue {
		suite.t.Error("Got: nothing | Expected: more than nothing")
	}
}

// AssertNotEmptyf tests if the value corresponds to its empty value, fail tests if it's the case.
func (suite Adapter) AssertNotEmptyf(got any, emptyValue any) {
	suite.t.Helper()
	if got == emptyValue {
		suite.t.Fatal("Got: nothing | Expected: more than nothing")
	}
}

// AssertErrIs tests if two errors are different, print an error if they are.
func (suite Adapter) AssertErrIs(got, expected error) {
	suite.t.Helper()
	if !errors.Is(got, expected) {
		suite.t.Errorf("Got: %v | Expected: %v\n", got, expected)
	}
}

// AssertErrIsf tests if two errors are different, fail tests if they are.
func (suite Adapter) AssertErrIsf(got, expected error) {
	suite.t.Helper()
	if !errors.Is(got, expected) {
		suite.t.Fatalf("Got: %v | Expected: %v\n", got, expected)
	}
}

// AssertErrAs tests if two errors are of different types, print an error if they are.
func (suite Adapter) AssertErrAs(got error, expected any) {
	suite.t.Helper()
	if !errors.As(got, &expected) {
		suite.t.Errorf("Got: %v | Expected: %v\n", got, expected)
	}
}

// AssertErrAsf tests if two errors are of different types, fail tests if they are.
func (suite Adapter) AssertErrAsf(got error, expected any) {
	suite.t.Helper()
	if !errors.As(got, &expected) {
		suite.t.Fatalf("Got: %v | Expected: %v\n", got, expected)
	}
}

// AssertErr tests if there is an error, print an error if it's not the case.
func (suite Adapter) AssertErr(expected error) {
	suite.t.Helper()
	suite.AssertErrIs(expected, expected)
}

// AssertErrf tests if there is an error, fail the tests if it's not the case.
func (suite Adapter) AssertErrf(expected error) {
	suite.t.Helper()
	suite.AssertErrIsf(expected, expected)
}

// AssertNoErr tests if there is no error, print an error if there's an error.
func (suite Adapter) AssertNoErr(got error) {
	suite.t.Helper()
	suite.AssertErrIs(got, nil)
}

// AssertNoErrf tests if there is no error, fail the tests if there's an error.
func (suite Adapter) AssertNoErrf(got error) {
	suite.t.Helper()
	suite.AssertErrIsf(got, nil)
}
