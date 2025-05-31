package server

import (
	crand "crypto/rand"
	"io"
	mrand "math/rand"
)

func simulateFile(n int64) ([]byte, int64, error) {
	size := mrand.Int63n(n)
	file := make([]byte, size)
	_, err := io.ReadFull(crand.Reader, file)
	if err != nil {
		return nil, 0, err
	}

	return file, size, nil
}
