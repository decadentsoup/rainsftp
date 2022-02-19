package main

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

type writer struct {
	context     context.Context
	log         *logrus.Entry
	tempFile    *os.File
	minioClient *minio.Client
	bucket      string
	key         string
}

func newWriter(context context.Context, log *logrus.Entry, minioClient *minio.Client, bucket string, key string) (*writer, error) {
	if tempFile, err := os.CreateTemp("", "rainsftp"); err != nil {
		return nil, err
	} else {
		return &writer{
			context:     context,
			log:         log,
			tempFile:    tempFile,
			minioClient: minioClient,
			bucket:      bucket,
			key:         key,
		}, nil
	}
}

func (writer *writer) WriteAt(buffer []byte, offset int64) (int, error) {
	return writer.tempFile.WriteAt(buffer, offset)
}

func (writer *writer) Close() error {
	if _, err := writer.minioClient.PutObject(writer.context, writer.bucket, writer.key, writer.tempFile, -1, minio.PutObjectOptions{}); err != nil {
		writer.cleanUp()
		return err
	}

	return writer.cleanUp()
}

func (writer *writer) cleanUp() error {
	name := writer.tempFile.Name()

	// We can ignore the result of this operation as long as os.Remove succeeds.
	// We do not care if the data was successfully commit to the filesystem.
	writer.tempFile.Close()

	if err := os.Remove(name); err != nil {
		writer.log.WithError(err).WithField("name", name).Warn("failed to delete temporary file")
	}

	return nil
}
