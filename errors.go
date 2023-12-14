//    reddlinks, a simple link shortener written in Go.
//    Copyright (C) 2023 redd
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

package main

import "errors"

var (
	errEmpty                = errors.New("can't be empty")
	errRead                 = errors.New("couldn't be read")
	errNotChecked           = errors.New("couldn't be checked")
	errInvalid              = errors.New("is invalid")
	errInvalidOrUnsupported = errors.New("is invalid or unsupported")
	errNullOrNegative       = errors.New("can't be null or negative")
	errSuperior             = errors.New("can't be superior to")
	errInferior             = errors.New("can't be inferior to")
)
