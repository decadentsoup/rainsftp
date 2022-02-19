package main

import (
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
		if object, err := handler.minioClient.GetObject(request.Context(), handler.bucket, strings.TrimPrefix(request.Filepath, "/"), minio.GetObjectOptions{}); err != nil {
			return nil, err
		} else {
			// Theoretically we should be able to just return object, but due to
			// EOF handling doing that will always result in an HTTP 416 getting
			// returned to the client once the download completes. The file will
			// arrive to the client, but the client will report an error.
			return reader{object}, nil
		}
	default:
		return nil, sftp.ErrSSHFxOpUnsupported
	}
}

func (handler handler) Filewrite(request *sftp.Request) (io.WriterAt, error) {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "Put":
		return newWriter(request.Context(), handler.log, handler.minioClient, handler.bucket, strings.TrimPrefix(request.Filepath, "/"))
	case "Open":
		return nil, sftp.ErrSSHFxOpUnsupported
	default:
		return nil, sftp.ErrSSHFxOpUnsupported
	}
}

func (handler handler) Filecmd(request *sftp.Request) error {
	handler.log.WithField("request", request).Debug("request received")

	switch request.Method {
	case "Setstat":
		return sftp.ErrSSHFxOpUnsupported
	case "Rename":
		return sftp.ErrSSHFxOpUnsupported
	case "Rmdir":
		return sftp.ErrSSHFxOpUnsupported
	case "Mkdir":
		return sftp.ErrSSHFxOpUnsupported
	case "Link":
		return sftp.ErrSSHFxOpUnsupported
	case "Symlink":
		return sftp.ErrSSHFxOpUnsupported
	case "Remove":
		return sftp.ErrSSHFxOpUnsupported
	default:
		return sftp.ErrSSHFxOpUnsupported
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
		return nil, sftp.ErrSSHFxOpUnsupported
	default:
		return nil, sftp.ErrSSHFxOpUnsupported
	}
}
