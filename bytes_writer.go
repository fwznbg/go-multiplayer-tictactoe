package main

import "bytes"

type BytesWriter struct {
	buffer *bytes.Buffer
}

func (bw *BytesWriter) Write(p []byte) (n int, err error) {
	return bw.buffer.Write(p)
}

func (bw *BytesWriter) Bytes() []byte {
	return bw.buffer.Bytes()
}

func NewBytesWriter() *BytesWriter {
	return &BytesWriter{
		buffer: &bytes.Buffer{},
	}
}
