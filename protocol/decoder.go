package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) Decode() (Value, error) {
	prefix, err := d.r.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch prefix {
	case byte(SimpleString):
		value, err := d.decodeSimpleString()
		return value, err
	case byte(Integer):
		return d.decodeInteger()
	case byte(Error):
		return d.decodeError()
	case byte(BulkString):
		return d.decodeBulkString()
	case byte(Array):
		return d.decodeArray()
	default:
		return Value{}, fmt.Errorf("invalid prefix")
	}
}

func (d *Decoder) decodeSimpleString() (Value, error) {
	readBytes, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	value := Value{
		Type: SimpleString,
		Str:  string(readBytes),
	}
	return value, nil
}

func (d *Decoder) decodeInteger() (Value, error) {
	readBytes, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	number, err := strconv.ParseInt(string(readBytes), 10, 64)
	if err != nil {
		return Value{}, err
	}
	value := Value{
		Type: Integer,
		Num:  number,
	}
	return value, nil
}

func (d *Decoder) decodeError() (Value, error) {
	readBytes, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{
		Type: Error,
		Str:  string(readBytes),
	}, nil
}

func (d *Decoder) decodeBulkString() (Value, error) {
	readBytes, err := d.readLine()
	if err != nil {
		return Value{}, err
	}
	length, err := strconv.ParseInt(string(readBytes), 10, 64)
	if err != nil {
		return Value{}, err
	}
	if length < -1 {
		return Value{}, fmt.Errorf("invalid length")
	}
	if length == -1 {
		return Value{
			Type: BulkString,
			Num:  -1,
		}, nil
	}
	buf := make([]byte, length)
	n, err := io.ReadFull(d.r, buf)
	if err != nil {
		return Value{}, err
	}
	delimBuf := make([]byte, 2)
	_, err = io.ReadFull(d.r, delimBuf)
	if err != nil {
		return Value{}, err
	}
	if delimBuf[0] != '\r' || delimBuf[1] != '\n' {
		return Value{}, ErrInvalidRESPDelim
	}
	return Value{
		Type: BulkString,
		Str:  string(buf[:n]),
	}, nil
}

func (d *Decoder) readLine() ([]byte, error) {
	readBytes, err := d.r.ReadBytes('\r')
	if err != nil {
		return nil, err
	}
	nextByte, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	if nextByte != '\n' {
		return nil, ErrInvalidRESPDelim
	}
	readBytes = readBytes[:len(readBytes)-1]
	return readBytes, nil
}

func (d *Decoder) decodeArray() (Value, error) {
	readBytes, err := d.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.ParseInt(string(readBytes), 10, 64)
	if err != nil {
		return Value{}, err
	}

	if length < -1 {
		return Value{}, fmt.Errorf("invalid array length")
	}

	if length == -1 {
		return Value{
			Type:  Array,
			Num:   -1,
			Array: nil,
		}, nil
	}

	values := make([]Value, 0, length)

	for i := int64(0); i < length; i++ {
		value, err := d.Decode()
		if err != nil {
			return Value{}, err
		}

		values = append(values, value)
	}

	return Value{
		Type:  Array,
		Num:   length,
		Array: values,
	}, nil
}
