package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
)

type handler struct {
	log         *logrus.Entry
	minioClient *minio.Client
	bucket      string
}

func (handler handler) Fileread(request *sftp.Request) (io.ReaderAt, error) {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "Get":
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("not implemented")
	}
}

func (handler handler) Filewrite(request *sftp.Request) (io.WriterAt, error) {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "Put":
		return nil, errors.New("not implemented")
	case "Open":
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("not implemented")
	}
}

func (handler handler) Filecmd(request *sftp.Request) error {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "Setstat":
		return errors.New("not implemented")
	case "Rename":
		return errors.New("not implemented")
	case "Rmdir":
		return errors.New("not implemented")
	case "Mkdir":
		return errors.New("not implemented")
	case "Link":
		return errors.New("not implemented")
	case "Symlink":
		return errors.New("not implemented")
	case "Remove":
		return errors.New("not implemented")
	default:
		return errors.New("not implemented")
	}
}

func (handler handler) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "List":
		fileName := strings.TrimPrefix(request.Filepath, "/")

		// If this is the root directory, then we want an empty string.
		// Otherwise, we want a string ending with "/".
		if len(fileName) > 0 {
			fileName = fileName + "/"
		}

		fileInfoList := make([]os.FileInfo, 0)

		for object := range handler.minioClient.ListObjects(request.Context(), handler.bucket, minio.ListObjectsOptions{Prefix: fileName}) {
			if object.Err != nil {
				return nil, object.Err
			}

			if object.Key != fileName {
				fileInfoList = append(fileInfoList, fileInfoFromObjectInfo(object))
			}
		}

		return listerAt(fileInfoList), nil
	case "Stat":
		fileName := strings.TrimPrefix(request.Filepath, "/")
		// Condition 1: fileName is "", we need to fake a root directory
		// Condition 2: object is a file, we can pass the fileName as-is
		// Condition 3: object is a directory, we must append "/" to the end
		if len(fileName) == 0 {
			return listerAt([]os.FileInfo{
				&fileInfo{name: "/", size: 0, modTime: time.Now(), isDir: true},
			}), nil
		} else if object, err := handler.minioClient.StatObject(request.Context(), handler.bucket, fileName, minio.GetObjectOptions{}); err == nil {
			return listerAt([]os.FileInfo{fileInfoFromObjectInfo(object)}), nil
		} else if object, err = handler.minioClient.StatObject(request.Context(), handler.bucket, fileName+"/", minio.GetObjectOptions{}); err == nil {
			return listerAt([]os.FileInfo{fileInfoFromObjectInfo(object)}), nil
		} else {
			return nil, err
		}
	case "Readlink":
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("not implemented")
	}
}
