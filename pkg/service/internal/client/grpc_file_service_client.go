package client

import (
	"context"
	"errors"
	"io"

	"github.com/Mikhalevich/filesharing/pkg/proto/file"
)

// GRPCFileServiceClient it's just wrapper around grpc FileServiceClient
type GRPCFileServiceClient struct {
	client file.FileService
}

// NewGRPCFileServiceClient create new client
func NewGRPCFileServiceClient(c file.FileService) *GRPCFileServiceClient {
	return &GRPCFileServiceClient{
		client: c,
	}
}

// Files return files from storage
func (c *GRPCFileServiceClient) Files(storage string, isPermanent bool) ([]*file.File, error) {
	rsp, err := c.client.List(context.Background(), &file.ListRequest{Storage: storage, IsPermanent: isPermanent})
	if err != nil {
		return nil, err
	}

	return rsp.GetFiles(), nil
}

// CreateStorage just create storage with specified storage name and permanent folder
func (c *GRPCFileServiceClient) Create(storage string, withPermanent bool) error {
	if _, err := c.client.CreateStorage(context.Background(), &file.CreateStorageRequest{
		Name:          storage,
		WithPermanent: withPermanent,
	}); err != nil {
		return err
	}
	return nil
}

// Remove remove file with fileName from storage
func (c *GRPCFileServiceClient) Remove(storage string, isPermanent bool, fileName string) error {
	_, err := c.client.RemoveFile(context.Background(), &file.FileRequest{
		Storage:     storage,
		IsPermanent: isPermanent,
		FileName:    fileName,
	})

	return err
}

// Get download file from storage
func (c *GRPCFileServiceClient) Get(storage string, isPermanent bool, fileName string, w io.Writer) error {
	stream, err := c.client.GetFile(context.Background(), &file.FileRequest{
		Storage:     storage,
		IsPermanent: isPermanent,
		FileName:    fileName,
	})
	if err != nil {
		return err
	}

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		_, err = w.Write(chunk.Content)
		if err != nil {
			return err
		}
	}

	return nil
}

// Upload upload file to storage
func (c *GRPCFileServiceClient) Upload(storage string, isPermanent bool, fileName string, r io.Reader) (*file.File, error) {
	stream, err := c.client.UploadFile(context.Background())
	if err != nil {
		return nil, err
	}

	stream.Send(&file.FileUploadRequest{
		FileChunk: &file.FileUploadRequest_Metadata{
			Metadata: &file.FileRequest{
				Storage:     storage,
				IsPermanent: isPermanent,
				FileName:    fileName,
			},
		},
	})

	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			stream.Send(&file.FileUploadRequest{
				FileChunk: &file.FileUploadRequest_Content{
					Content: buf[:n],
				},
			})
		}
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
	}

	err = stream.Send(&file.FileUploadRequest{
		FileChunk: &file.FileUploadRequest_End{
			End: true,
		},
	})

	if err != nil {
		return nil, err
	}

	f, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return f, nil
}
