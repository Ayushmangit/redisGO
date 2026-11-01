package protocol

import "errors"

type ValueType byte

var ErrInvalidRESPDelim = errors.New("invalid RESP delim")

const (
	SimpleString ValueType = '+'
	Error        ValueType = '-'
	Integer      ValueType = ':'
	BulkString   ValueType = '$'
	Array        ValueType = '*'
)

type Value struct {
	Type  ValueType
	Str   string
	Num   int64
	Array []Value
}
