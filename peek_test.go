package peek

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

const (
	size = 100 * 1024 * 1024
)

func isGzip(input io.Reader) (io.Reader, bool, error) {
	buf := [3]byte{}

	n, err := io.ReadAtLeast(input, buf[:], len(buf))
	if err != nil {
		return nil, false, err
	}

	isGzip := buf[0] == 0x1F && buf[1] == 0x8B && buf[2] == 0x8
	return io.MultiReader(bytes.NewReader(buf[:n]), input), isGzip, nil
}

var head = []byte{0x1F, 0x8B, 0x8}

func isGzip2(input io.Reader) (io.Reader, bool, error) {
	r := bufio.NewReader(input)
	buf, err := r.Peek(len(head))
	if err != nil {
		return nil, false, err
	}
	return r, bytes.Equal(buf, head), nil
}

func BenchmarkIsGzip(b *testing.B)  { bench(b, isGzip) }
func BenchmarkIsGzip2(b *testing.B) { bench(b, isGzip2) }

func bench(b *testing.B, f func(io.Reader) (io.Reader, bool, error)) {
	data := bytes.Repeat([]byte("A"), size)
	w := ioutil.Discard
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(data)
		read, gz, err := f(r)
		if err != nil {
			b.Fatal(err)
		}
		if gz {
			b.Fatal("is gzip")
		}
		n, err := io.Copy(w, read)
		if err != nil {
			b.Fatal(err)
		}
		if n != size {
			b.Fatal("n!=size")
		}
	}
}

func TestWriteTo(t *testing.T) {
	fn := func(f func(io.Reader) (io.Reader, bool, error)) {
		data := bytes.Repeat([]byte("A"), size)
		w := ioutil.Discard
		r := bytes.NewBuffer(data)
		read, _, _ := f(r)
		if _, ok := w.(io.ReaderFrom); ok {
			t.Log("has ReaderFrom")
		}
		if _, ok := read.(io.WriterTo); ok {
			t.Log("has WriterTo")
		}
	}
	t.Run("isGzip", func(t *testing.T) {
		fn(isGzip)
	})
	t.Run("isGzip2", func(t *testing.T) {
		fn(isGzip2)
	})
}
