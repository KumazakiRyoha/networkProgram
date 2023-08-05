package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

type Binary []byte

func (m Binary) Bytes() []byte  { return m }
func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(w io.Writer) (int64, error) {
	// 这个操作相当于发送了一个“我要发送的数据是二进制类型”的信号。
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1-byte-type
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	// 这个操作相当于发送了一个“我要发送的数据的长度是多少”的信号。
	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	n += 4
	// 这个操作发送实际的值 TLV(Type-Length-Value)编码
	o, err := w.Write(m) // payload
	return n + int64(o), err
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// binary.Read会读取与&typ相同大小的数据。这里&typ是一个uint8类型的
	// 指针，其大小为1字节
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-size-byte
	if err != nil {
		return n, err
	}
	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}
	*m = make([]byte, size)
	o, err := r.Read(*m)
	return n + int64(o), err
}

type String string

func (m String) Bytes() []byte {
	return []byte(m)
}
func (m String) String() string {
	return string(m)
}

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) //	1-byte type
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}

	n += 4
	o, err := w.Write([]byte(m)) // payload
	return n + int64(o), err

}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// binary.Read 函数会从 r 中读取数据，然后将这些数据解码并存储到 &typ
	// 指向的内存位置中。因为 &typ 是一个指针，所以 binary.Read 可以修改它指向的值。
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	if typ != StringType {
		return n, errors.New("invalid String")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	n += 4
	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}
	*m = String(buf)

	return n + int64(o), nil

}

func decode(r io.Reader) (Payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}

	/**
	在 Go 的 io.Reader 中，读取操作是顺序进行的，并且每读取一次，读取的位置就会向后移动。
	这意味着如果你已经从 Reader 中读取了一个字节，那么下次再进行读取时，你将从上次读取结束的位置开始，
	而不是从最开始的位置。因此，一旦一个字节被读取，它就不能再被重新读取，除非你显式地将读取的位置移动回去。
	io.MultiReader 可以将多个 Reader 连接在一起，并创建一个新的 Reader。当从新的 Reader 中读取数据时，
	它首先从第一个 Reader 中读取数据，当第一个 Reader 的数据读取完后，再从第二个 Reader 中读取数据。因此，
	你可以将一个只包含类型的 Reader 和原来的 Reader 连接在一起，这样 ReadFrom 方法就可以按照预期的顺序读取字节了。
	*/
	_, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, nil
}
