package goannoy

import "errors"

var (
	ErrInvalidDimension    = errors.New("invalid dimension")
	ErrInvalidVector       = errors.New("invalid vector")
	ErrInvalidID           = errors.New("invalid item ID")
	ErrInvalidSearchK      = errors.New("invalid search budget")
	ErrInvalidOption       = errors.New("invalid option")
	ErrClosed              = errors.New("index is closed")
	ErrCorruptIndex        = errors.New("corrupt index")
	ErrIncompatibleIndex   = errors.New("incompatible index")
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)
