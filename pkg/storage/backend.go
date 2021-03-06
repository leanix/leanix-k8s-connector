package storage

import (
	"errors"
	"fmt"
)

const (
	// AzureBlobStorage is a constant for the azure blob storage identifier
	AzureBlobStorage string = "azureblob"
	// FileStorage is a constant for the file storage identifier
	FileStorage string = "file"
	// LdifFileName is a constant for the file name used to store the ldif content
	LdifFileName string = "kubernetes.ldif"
	// LogFileName is a constant for the file name used to store the log output
	LogFileName string = "leanix-k8s-connector.log"
)

// Backend exposes a common interface for all storage mechanisms
type Backend interface {
	UploadLdif(ldif []byte) error
	UploadLog(log []byte) error
}

// NewBackend create a new storage backend for the given storage backend type
func NewBackend(backend string, azureOpts *AzureBlobOpts, localFileOpts *LocalFileOpts) (Backend, error) {
	switch backend {
	case AzureBlobStorage:
		if azureOpts == nil {
			return nil, errors.New("azure storage options must be set when using azure as storage target")
		}
		return NewAzureBlob(azureOpts)
	case FileStorage:
		return NewLocalFile(localFileOpts.Path)
	}
	return nil, fmt.Errorf("Unsupported storage backend type %s", backend)
}
