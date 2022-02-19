package main

import (
	"github.com/minio/minio-go/v7"
)

type reader struct{ object *minio.Object }

func (reader reader) ReadAt(buffer []byte, offset int64) (int, error) {
	if n, err := reader.object.ReadAt(buffer, offset); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}
