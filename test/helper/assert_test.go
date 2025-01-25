//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2025 redd
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

package helper_test

import (
	"errors"
	"testing"

	"github.com/redds-be/reddlinks/test/helper"
)

// mockStruct is a struct for a mock struct.
type mockStruct struct {
	exists bool
}

// mockError is a struct for a mock error.
type mockError struct{}

// Error returns a mock error (for TestAssertErrAs).
func (m *mockError) Error() string {
	return "this is an error"
}

// willReturnError will return an error (for TestAssertErrAs).
func willReturnError() error {
	return &mockError{}
}

// The test suite struct.
type assertTestSuite struct {
	t *testing.T
	a helper.Adapter
}

// TestAssertBool tests assert with booleans.
func (suite assertTestSuite) TestAssertBool() {
	suite.a.Assert(true, true)
	suite.a.Assertf(true, true)
}

// TestAssertInt tests assert with int.
func (suite assertTestSuite) TestAssertInt() {
	suite.a.Assert(1, 1)
	suite.a.Assert(-10, -10)
	suite.a.Assertf(1, 1)
	suite.a.Assertf(-10, -10)
}

// TestAssertInt8 tests assert with int8.
func (suite assertTestSuite) TestAssertInt8() {
	suite.a.Assert(int8(1), int8(1))
	suite.a.Assert(int8(1), int8(1))
	suite.a.Assertf(int8(-10), int8(-10))
	suite.a.Assertf(int8(-10), int8(-10))
}

// TestAssertInt16 tests assert with int16.
func (suite assertTestSuite) TestAssertInt16() {
	suite.a.Assert(int16(1), int16(1))
	suite.a.Assert(int16(-10), int16(-10))
	suite.a.Assertf(int16(1), int16(1))
	suite.a.Assertf(int16(-10), int16(-10))
}

// TestAssertInt32 tests assert with int32.
func (suite assertTestSuite) TestAssertInt32() {
	suite.a.Assert(int32(1), int32(1))
	suite.a.Assert(int32(1), int32(1))
	suite.a.Assertf(int32(-10), int32(-10))
	suite.a.Assertf(int32(-10), int32(-10))
}

// TestAssertInt64 tests assert with int64.
func (suite assertTestSuite) TestAssertInt64() {
	suite.a.Assert(int64(1), int64(1))
	suite.a.Assert(int64(1), int64(1))
	suite.a.Assertf(int64(-10), int64(-10))
	suite.a.Assertf(int64(-10), int64(-10))
}

// TestAssertUint tests assert with uint.
func (suite assertTestSuite) TestAssertUint() {
	suite.a.Assert(uint(1), uint(1))
	suite.a.Assertf(uint(1), uint(1))
}

// TestAssertUint8 tests assert with uint8.
func (suite assertTestSuite) TestAssertUint8() {
	suite.a.Assert(uint8(1), uint8(1))
	suite.a.Assertf(uint8(1), uint8(1))
}

// TestAssertUint16 tests assert with uint16.
func (suite assertTestSuite) TestAssertUint16() {
	suite.a.Assert(uint16(1), uint16(1))
	suite.a.Assertf(uint16(1), uint16(1))
}

// TestAssertUint32 tests assert with uint32.
func (suite assertTestSuite) TestAssertUint32() {
	suite.a.Assert(uint32(1), uint32(1))
	suite.a.Assertf(uint32(1), uint32(1))
}

// TestAssertUint64 tests assert with uint64.
func (suite assertTestSuite) TestAssertUint64() {
	suite.a.Assert(uint64(1), uint64(1))
	suite.a.Assertf(uint64(1), uint64(1))
}

// TestAssertUintptr tests assert with uintptr.
func (suite assertTestSuite) TestAssertUintptr() {
	suite.a.Assert(uintptr(1), uintptr(1))
	suite.a.Assertf(uintptr(1), uintptr(1))
}

// TestAssertFloat32 tests assert with float32.
func (suite assertTestSuite) TestAssertFloat32() {
	suite.a.Assert(float32(1.5), float32(1.5))
	suite.a.Assertf(float32(1.5), float32(1.5))
}

// TestAssertFloat64 tests assert with float64.
func (suite assertTestSuite) TestAssertFloat64() {
	suite.a.Assert(1.5, 1.5)
	suite.a.Assertf(1.5, 1.5)
}

// TestAssertComplex64 tests assert with complex64.
func (suite assertTestSuite) TestAssertComplex64() {
	suite.a.Assert(complex64(1), complex64(1))
	suite.a.Assertf(complex64(1), complex64(1))
}

// TestAssertComplex128 tests assert with complex128.
func (suite assertTestSuite) TestAssertComplex128() {
	suite.a.Assert(complex128(1), complex128(1))
	suite.a.Assertf(complex128(1), complex128(1))
}

// TestAssertStr tests assert with strings.
func (suite assertTestSuite) TestAssertStr() {
	suite.a.Assert("String", "String")
	suite.a.Assertf("String", "String")
}

// TestAssertNotEmpty tests assertNotEmpty using mockStuct.
func (suite assertTestSuite) TestAssertNotEmpty() {
	mock := mockStruct{exists: true}
	suite.a.AssertNotEmpty(mock, mockStruct{})
	suite.a.AssertNotEmptyf(mock, mockStruct{})
}

// TestAssertErrIs tests assert with errors.
func (suite assertTestSuite) TestAssertErrIs() {
	err := errors.New("this is an error") //nolint:goerr113
	suite.a.AssertErrIs(err, err)
	suite.a.AssertErrIsf(err, err)
}

// TestAssertErrAs tests assert with errors.
func (suite assertTestSuite) TestAssertErrAs() {
	err := willReturnError()
	var falseErr *mockError
	suite.a.AssertErrAs(err, &falseErr)
	suite.a.AssertErrAsf(err, &falseErr)
}

// TestAssertErr tests assert if there's an error.
func (suite assertTestSuite) TestAssertErr() {
	err := willReturnError()
	suite.a.AssertErr(err)
	suite.a.AssertErrf(err)
}

// TestAssertErr tests assert if there's no error.
func (suite assertTestSuite) TestAssertNoErr() {
	suite.a.AssertNoErr(nil)
	suite.a.AssertNoErrf(nil)
}

// Run tests for assert.go.
func TestAssert(t *testing.T) {
	// Run tests in parallel
	t.Parallel()

	assertAdapter := helper.NewAdapter(t)

	// Initialize the test suite
	suite := assertTestSuite{t: t, a: assertAdapter}

	// Run the tests for pretty much all types
	suite.TestAssertBool()
	suite.TestAssertInt()
	suite.TestAssertInt8()
	suite.TestAssertInt16()
	suite.TestAssertInt32()
	suite.TestAssertInt64()
	suite.TestAssertUint()
	suite.TestAssertUint8()
	suite.TestAssertUint16()
	suite.TestAssertUint32()
	suite.TestAssertUint64()
	suite.TestAssertUintptr()
	suite.TestAssertFloat32()
	suite.TestAssertFloat64()
	suite.TestAssertComplex64()
	suite.TestAssertComplex128()
	suite.TestAssertStr()
	suite.TestAssertNotEmpty()
	suite.TestAssertErrIs()
	suite.TestAssertErrAs()
	suite.TestAssertErr()
	suite.TestAssertNoErr()
}
