package goson

import (
	"errors"
	"strconv"
)

type Type int32

const (
	NULL Type = iota
	FALSE
	TRUE
	NUMBER
	STRING
	ARRAY
	OBJECT
)

var (
	ErrParseInvalidValue             = errors.New("parse invalid value")
	ErrParseRootNotSingular          = errors.New("parse root not singular")
	ErrParseNumberTooBig             = strconv.ErrRange
	ErrParseInvalidStringEscape      = errors.New("parse invalid string escape")
	ErrParseMissQuotationMark        = errors.New("parse miss quotation mark")
	ErrParseInvalidStringChar        = errors.New("parse invalid string char")
	ErrParseInvalidUnicodeHex        = errors.New("parse invalid unicode hex")
	ErrParseInvalidUnicodeSurrogate  = errors.New("parse invalid unicode surrogate")
	ErrParseMissCommaOrSquareBracket = errors.New("parse miss comma or square bracket")
	ErrParseMissKey                  = errors.New("parse miss key")
	ErrParseMissColon                = errors.New("parse miss colon")
	ErrParseMissCommaOrCurlyBracket  = errors.New("parse miss comma or curly bracket")
	ErrKeyNotExist                   = errors.New("key not exist")
)
