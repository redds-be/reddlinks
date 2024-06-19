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

package env

import "errors"

// Define all the errors for the env package.
//
// ErrEmpty defines an error for empty variables,
// ErrRead defines an error for variables that couldn't be read,
// ErrNotChecked defines an error for variables where a value couldn't be checked,
// ErrInvalid defines an error for variables where a value is invalid,
// ErrInvalidOrUnsupported defines an error for variables where a value is either invalid or unsupported,
// ErrNullOrNegative defines an error for variables where a value is either null or negative,
// ErrNegative defines an error for variables where a value is negative,
// ErrSuperior defines an error for variables where a value can't be superior to another one,
// ErrInferior defines an error for variables where a value can't be inferior to another one.
var (
	ErrEmpty                = errors.New("can't be empty")
	ErrRead                 = errors.New("couldn't be read")
	ErrNotChecked           = errors.New("couldn't be checked")
	ErrInvalid              = errors.New("is invalid")
	ErrInvalidOrUnsupported = errors.New("is invalid or unsupported")
	ErrNullOrNegative       = errors.New("can't be null or negative")
	ErrNegative             = errors.New("can't be negative")
	ErrSuperior             = errors.New("can't be superior to")
	ErrInferior             = errors.New("can't be inferior to")
)
