package log

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// number of bytes used for storing record's length, bascillay size of a record struct
const (
	lenWidth = 8
)

// store struct to have a pointer to a file, bufio writer
type store struct {
	*os.File
	mu  sync.Mutex
	buf *bufio.Writer
}

// newStore returns a pointer to a new store struct for given file
func newStore(f *os.File) (*store, error) {
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
	}, nil

}

// Append function appends a p bytes into the buffer by using the buf field in the store struct
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	written, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	fmt.Fprintf(s.buf, "\n")
	err = s.buf.Flush()
	if err != nil {
		return 0, 0, err
	}
	return uint64(written), pos, nil
}

// Close Method
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
