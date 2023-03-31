package cloudstorage

import (
	"context"
	"errors"
	netUrl "net/url"
	"os"
	"path"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"

	util "github.com/bondhan/golib/util"
)

const (
	Host = "https://storage.googleapis.com/"
)

type clientCloudStorage struct {
	BucketName string
	FolderName string
}

type ClientCloudStorage interface {
	UploadFile(ctx context.Context, file string) (string, error)
	UploadFileWPath(ctx context.Context, path, file string) (string, error)
	DownloadFile(ctx context.Context, url string) (string, error)
	DownloadFileWPath(ctx context.Context, path, url string) (string, error)
}

func NewClientCloudStorage(bucketName, folderName string) (ClientCloudStorage, error) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, errors.New("GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	return &clientCloudStorage{
		BucketName: bucketName,
		FolderName: folderName,
	}, nil
}

func (c *clientCloudStorage) upload(ctx context.Context, in *blob.Bucket, file string) (string, error) {

	reader, err := in.NewReader(ctx, file, nil)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// write file to google cloud storage
	out, err := blob.OpenBucket(ctx, "gs://"+c.BucketName)
	if err != nil {
		return "", err
	}

	rootFolder := ""
	if c.FolderName != "" {
		out = blob.PrefixedBucket(out, c.FolderName+"/")
		rootFolder = c.FolderName + "/"
	}

	newName := strings.Join([]string{util.NewRandomString(10), "_", file}, "")
	writer, err := out.NewWriter(ctx, newName, nil)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	_, err = writer.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	url := strings.Join([]string{Host, c.BucketName, "/", rootFolder, newName}, "")
	return url, nil
}

func (c *clientCloudStorage) UploadFile(ctx context.Context, file string) (string, error) {
	// read file from local
	in, err := blob.OpenBucket(ctx, "file://")
	if err != nil {
		return "", err
	}

	return c.upload(ctx, in, file)
}

func (c *clientCloudStorage) UploadFileWPath(ctx context.Context, path, file string) (string, error) {
	// read file from local
	in, err := blob.OpenBucket(ctx, "file://"+path)
	if err != nil {
		return "", err
	}
	return c.upload(ctx, in, file)
}

func (c *clientCloudStorage) download(ctx context.Context, out *blob.Bucket, url string) (string, error) {

	// read file from google cloud storage
	in, err := blob.OpenBucket(ctx, "gs://"+c.BucketName)
	if err != nil {
		return "", err
	}

	if c.FolderName != "" {
		in = blob.PrefixedBucket(in, c.FolderName+"/")
	}

	unescapedPath, err := netUrl.PathUnescape(url)
	if err != nil {
		return "", err
	}
	fileName := path.Base(unescapedPath)

	reader, err := in.NewReader(ctx, fileName, nil)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	writer, err := out.NewWriter(ctx, fileName, nil)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	_, err = writer.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func (c *clientCloudStorage) DownloadFile(ctx context.Context, url string) (string, error) {

	// write file to local
	out, err := blob.OpenBucket(ctx, "file://")
	if err != nil {
		return "", err
	}

	return c.download(ctx, out, url)
}

func (c *clientCloudStorage) DownloadFileWPath(ctx context.Context, path, url string) (string, error) {
	// write file to local
	out, err := blob.OpenBucket(ctx, "file://"+path)
	if err != nil {
		return "", err
	}

	return c.download(ctx, out, url)
}
