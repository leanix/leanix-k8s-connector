package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// LocalFileOpts options for local file storage
type LocalFileOpts struct {
	Path string
}

// LocalFile writes the content to disk
type LocalFile struct {
	Path string
}

// NewLocalFile create a LocalFile StorageBackend
func NewLocalFile(path string) (*LocalFile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}
	lf := &LocalFile{
		Path: path,
	}
	return lf, nil
}

// Upload persists the ldif content in a local files
func (u *LocalFile) UploadLdif(ldif []byte, LdifFileName string) error {
	err := ioutil.WriteFile(path.Join(u.Path, LdifFileName), ldif, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Upload persists the the log file content in a local files
func (u *LocalFile) UploadLog(log []byte, LogFileName string) error {
	err := ioutil.WriteFile(path.Join(u.Path, LogFileName), log, 0644)
	if err != nil {
		return err
	}
	return nil
}
