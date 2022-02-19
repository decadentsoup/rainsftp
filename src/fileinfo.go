package main

import (
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

const fileMode = os.FileMode(0644)
const directoryMode = os.FileMode(0755) | os.ModeDir

func fileInfoFromObjectInfo(objectInfo minio.ObjectInfo) *fileInfo {
	return &fileInfo{
		name:    path.Base(objectInfo.Key),
		size:    objectInfo.Size,
		modTime: objectInfo.LastModified,
		isDir:   strings.HasSuffix(objectInfo.Key, "/"),
	}
}

func (fileInfo *fileInfo) Name() string       { return fileInfo.name }
func (fileInfo *fileInfo) Size() int64        { return fileInfo.size }
func (fileInfo *fileInfo) ModTime() time.Time { return fileInfo.modTime }
func (fileInfo *fileInfo) IsDir() bool        { return fileInfo.isDir }
func (fileInfo *fileInfo) Sys() interface{}   { return nil }

func (fileInfo *fileInfo) Mode() fs.FileMode {
	if fileInfo.isDir {
		return directoryMode
	} else {
		return fileMode
	}
}
