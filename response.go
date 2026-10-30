package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ResponseType int

var ErrInvalidResponseType = errors.New("invalid response type")

const (
	SimpleString ResponseType = iota
	ErrorResponse
	Integer
	BulkString
	NullBulkString
	Array
)

const (
	simpleStringHeader  = "+"
	errorResponseHeader = "-"
	integerHeader       = ":"
	bulkStringHeader    = "$"
	arrayHeader         = "*"

	registeredNurse = "\r\n"

	emptyArrayResponse = "*0\r\n"
	nullBulkString     = "$-1\r\n"
)

type Response struct {
	responseType ResponseType
	val          string
	intVal       int64
	values       []Response
}

func NewSimpleString(value string) *Response {
	return &Response{
		responseType: SimpleString,
		val:          value,
	}
}

func NewError(message string) *Response {
	return &Response{
		responseType: ErrorResponse,
		val:          message,
	}
}

func NewInteger(value int64) *Response {
	return &Response{
		responseType: Integer,
		intVal:       value,
	}
}

func NewBulkString(value string) *Response {
	return &Response{
		responseType: BulkString,
		val:          value,
	}
}

func NewNullBulkString() *Response {
	return &Response{
		responseType: NullBulkString,
	}
}

func NewArray(values []Response) *Response {
	return &Response{
		responseType: Array,
		values:       values,
	}
}

func (r Response) Encode() ([]byte, error) {
	res := ""
	switch r.responseType {
	case SimpleString:
		res = fmt.Sprintf("%s%v%s", simpleStringHeader, r.val, registeredNurse)
		return []byte(res), nil
	case ErrorResponse:
		res += errorResponseHeader
		res += r.val
		res += registeredNurse
		return []byte(res), nil
	case Integer:
		res += integerHeader
		res += strconv.FormatInt(r.intVal, 10)
		res += registeredNurse
		return []byte(res), nil
	case BulkString:
		res = encodeBulkString(r.val)
		return []byte(res), nil
	case NullBulkString:
		return []byte(nullBulkString), nil
	case Array:
		if len(r.values) == 0 {
			return []byte(emptyArrayResponse), nil
		}
		var builder strings.Builder
		fmt.Fprintf(&builder, "%s%d%s", arrayHeader, len(r.values), registeredNurse)
		for _, v := range r.values {
			bytes, err := v.Encode()
			if err != nil {
				return nil, err
			}
			builder.Write(bytes)
		}
		return []byte(builder.String()), nil
	default:
		return nil, ErrInvalidResponseType
	}
}

func encodeBulkString(value string) string {
	return fmt.Sprintf("%s%d%s%s%s", bulkStringHeader, len(value), registeredNurse, value, registeredNurse)
}
