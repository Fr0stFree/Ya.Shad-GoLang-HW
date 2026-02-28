//go:build !solution

package otp

import (
	"io"
)

type OTPReader struct {
	r    io.Reader
	prng io.Reader
	key  []byte
}

func (o *OTPReader) Read(p []byte) (n int, err error) {
	n, err = o.r.Read(p)
	if n <= 0 {
		return n, err
	}
	if cap(o.key) < n {
		o.key = make([]byte, n)
	}
	o.key = o.key[:n]

	_, _ = io.ReadFull(o.prng, o.key)

	for i := 0; i < n; i++ {
		p[i] ^= o.key[i]
	}
	return n, err
}

type OTPWriter struct {
	w      io.Writer
	prng   io.Reader
	buffer []byte
}

func (o *OTPWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if cap(o.buffer) < len(p) {
		o.buffer = make([]byte, len(p))
	}
	o.buffer = o.buffer[:len(p)]

	_, err = io.ReadFull(o.prng, o.buffer)
	if err != nil {
		return 0, err
	}
	for i := 0; i < len(p); i++ {
		o.buffer[i] ^= p[i]
	}
	return o.w.Write(o.buffer)
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &OTPReader{r: r, prng: prng, key: make([]byte, 1024)}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &OTPWriter{w: w, prng: prng}
}
