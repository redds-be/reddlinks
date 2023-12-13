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
