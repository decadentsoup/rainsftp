package main

import (
	"io"
	"os"
)

type listerAt []os.FileInfo

func (listerAt listerAt) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	n := 0

	if offset >= int64(len(listerAt)) {
		return 0, io.EOF
	}

	n = copy(ls, listerAt[offset:])

	if n < len(ls) {
		return n, io.EOF
	}

	return n, nil
}
