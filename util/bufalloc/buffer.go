// Copyright 2019 The Vearch Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package bufalloc

import (
	"io"
)

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
type Buffer interface {
	// Alloc allocs n bytes of slice from the buffer, growing the buffer as needed.
	// If n is negative, Alloc will panic. If the buffer can't grow it will panic with bytes.ErrTooLarge.
	Alloc(n int) []byte
	// Truncate discards all but the first n unread bytes from the buffer.
	// It panics if n is negative or greater than the length of the buffer.
	Truncate(n int)
	// Grow grows the buffer's capacity, if necessary, to guarantee space for n bytes.
	// If n is negative, Grow will panic. If the buffer can't grow it will panic with bytes.ErrTooLarge.
	Grow(n int)
	// WriteString appends the contents of s to the buffer, growing the buffer as
	// needed. The return value n is the length of s; err is always nil. If the
	// buffer becomes too large, WriteString will panic with ErrTooLarge.
	WriteString(s string) (n int, err error)
	// Write appends the contents of p to the buffer, growing the buffer as needed.
	// The return value n is the length of p; err is always nil.
	// If the buffer becomes too large, Write will panic with bytes.ErrTooLarge.
	Write(p []byte) (n int, err error)
	// WriteByte appends the byte c to the buffer, growing the buffer as needed.
	// If the buffer becomes too large, WriteByte will panic with bytes.ErrTooLarge.
	WriteByte(c byte) error
	// WriteTo writes data to w until the buffer is drained or an error occurs.
	// The return value n is the number of bytes written;
	// Any error encountered during the write is also returned.
	WriteTo(w io.Writer) (n int64, err error)
	// Read reads the next len(p) bytes from the buffer or until the buffer is drained.
	// The return value n is the number of bytes read.
	// If the buffer has no data to return, err is io.EOF (unless len(p) is zero); otherwise it is nil.
	Read(p []byte) (n int, err error)
	// ReadByte reads and returns the next byte from the buffer.
	// If no byte is available, it returns error io.EOF.
	ReadByte() (c byte, err error)
	// ReadBytes reads until the first occurrence of delim in the input,
	// returning a slice containing the data up to and including the delimiter.
	// If ReadBytes encounters an error before finding a delimiter, it returns the data read before the error and the error itself (often io.EOF).
	// ReadBytes returns err != nil if and only if the returned data does not end in delim.
	ReadBytes(delim byte) (line []byte, err error)
	// ReadFrom reads data from r until EOF and appends it to the buffer, growing the buffer as needed.
	// The return value n is the number of bytes read. Any error except io.EOF encountered during the read is also returned.
	// If the buffer becomes too large, ReadFrom will panic with bytes.ErrTooLarge.
	ReadFrom(r io.Reader) (n int64, err error)
	// Bytes returns a slice of the contents of the unread portion of the buffer;
	// If the caller changes the contents of the returned slice, the contents of the buffer will change,
	// provided there are no intervening method calls on the Buffer.
	Bytes() []byte
	// Next returns a slice containing the next n bytes from the buffer, advancing the buffer as if the bytes had been returned by Read.
	// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
	// The slice is only valid until the next call to a read or write method.
	Next(n int) []byte
	// Reset resets the buffer so it has no content.
	// b.Reset() is the same as b.Truncate(0).
	Reset()
	// String returns the contents of the unread portion of the buffer as a string.
	// If the Buffer is a nil pointer, it returns "<nil>".
	String() string
	// Len returns the number of bytes of the unread portion of the buffer;
	Len() int
	// Cap returns the capacity of the buffer.
	Cap() int
}

// AllocBuffer allocation byte buffer to a given size
func AllocBuffer(n int) Buffer {
	return buffPool.getBuffer(n)
}

// FreeBuffer release buffer to pool
func FreeBuffer(buf Buffer) {
	if buf == nil {
		return
	}
	buffPool.putBuffer(buf)
}
