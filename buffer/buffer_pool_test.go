package buffer

import (
	/*	"io"*/
	"bytes"
	"testing"
)

func TestBufferPool(t *testing.T) {
	pool := NewPool(10, 1024)

	text := []byte("Hello World")

	// get one
	bpe := pool.Take(bytes.NewBuffer(text))
	// read all
	out := make([]byte, 1024)
	l, _ := bpe.Br.Read(out)
	if bytes.Compare(out[0:l], text) != 0 {
		t.Errorf("Read invalid data: %v", out)
	}
	if l != len(text) {
		t.Errorf("Expected length %v got %v", len(text), l)
	}
	pool.Give(bpe)

	// get one
	bpe = pool.Take(bytes.NewBuffer(text))
	// read all
	out = make([]byte, 1024)
	l, _ = bpe.Br.Read(out)
	if bytes.Compare(out[0:l], text) != 0 {
		t.Errorf("Read invalid data: %v", out)
	}
	if l != len(text) {
		t.Errorf("Expected length %v got %v", len(text), l)
	}
	pool.Give(bpe)

	// get one
	bpe = pool.Take(bytes.NewBuffer(text))
	// read 1 byte
	out = make([]byte, 1)
	bpe.Br.Read(out)
	pool.Give(bpe)

	// get one
	bpe = pool.Take(bytes.NewBuffer(text))
	// read all
	out = make([]byte, 1024)
	l, _ = bpe.Br.Read(out)
	if bytes.Compare(out[0:l], text) != 0 {
		t.Errorf("Read invalid data: %v", out)
	}
	if l != len(text) {
		t.Errorf("Expected length %v got %v", len(text), l)
	}
	pool.Give(bpe)

}
