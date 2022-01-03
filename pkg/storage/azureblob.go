package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// AzureBlobOpts options for azure blob storage
type AzureBlobOpts struct {
	AccountName string
	AccountKey  string
	Container   string
}

// AzureContainer is used to create containers and upload files to Azure blob storage
type AzureContainer struct {
	Container *azblob.ContainerURL
}

// NewAzureBlob creates a new AzureBlob
func NewAzureBlob(azureOpts *AzureBlobOpts) (*AzureContainer, error) {
	if azureOpts == nil {
		return nil, errors.New("missing azure options")
	}

	credential, err := azblob.NewSharedKeyCredential(azureOpts.AccountName, azureOpts.AccountKey)
	if err != nil {
		panic(err)
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	URL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", azureOpts.AccountName, azureOpts.Container))
	if err != nil {
		return nil, err
	}
	container := azblob.NewContainerURL(*URL, pipeline)

	ctx := context.Background()
	_, err = container.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	// err is ignored, as the container can already exists

	u := &AzureContainer{
		Container: &container,
	}

	return u, nil
}

// Upload uploads the LDIF file to azure blob storage
func (u *AzureContainer) UploadLdif(ldif []byte, LdifFileName string) error {
	err := u.uploadFile(LdifFileName, ldif)
	if err != nil {
		return err
	}

	return nil
}

// Upload uploads the log file to azure blob storage
func (u *AzureContainer) UploadLog(log []byte, LogFileName string) error {
	err := u.uploadFile(LogFileName, log)
	if err != nil {
		return err
	}

	return nil
}

func (u *AzureContainer) uploadFile(name string, content []byte) error {
	blobURL := azblob.ContainerURL(*u.Container).NewBlockBlobURL(name)

	ctx := context.Background()
	_, err := azblob.UploadBufferToBlockBlob(ctx, content, blobURL, azblob.UploadToBlockBlobOptions{})

	return err
}

func UploadFileToSasUrl(blobSasUrl url.URL, content []byte) error {
	ctx := context.Background()
	_, err := azblob.UploadBufferToBlockBlob(ctx, content, azblob.NewBlockBlobURL(blobSasUrl, azblob.NewPipeline(azblob.NewAnonymousCredential(), azblob.PipelineOptions{})), azblob.UploadToBlockBlobOptions{})
	if err != nil {
		return err
	}
	return nil
}

func UploadConnectorLog(connectorLoggingUrl string, logs []byte) (string, error) {
	parse, err := url.Parse(connectorLoggingUrl)
	if err != nil {
		return "", err
	}
	err = UploadFileToSasUrl(*parse, logs)
	if err != nil {
		return "", err
	}
	return "", nil
}
