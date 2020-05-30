package main

import (
	"context"
	"errors"
	"io"

	"github.com/Mikhalevich/file_service/proto"
	"github.com/Mikhalevich/filesharing/handlers"
)

// GRPCFileServiceClient it's just wrapper around grpc FileServiceClient
type GRPCFileServiceClient struct {
	client proto.FileService
}

// NewGRPCFileServiceClient create new client
func NewGRPCFileServiceClient(c proto.FileService) *GRPCFileServiceClient {
	return &GRPCFileServiceClient{
		client: c,
	}
}

func unmarshalFile(file *proto.File) *handlers.File {
	return &handlers.File{
		Name:    file.GetName(),
		Size:    file.GetSize(),
		ModTime: file.GetModTime(),
	}
}

// Files return files from storage
func (c *GRPCFileServiceClient) Files(storage string, isPermanent bool) ([]*handlers.File, error) {
	r, err := c.client.List(context.Background(), &proto.ListRequest{Storage: storage, IsPermanent: isPermanent})
	if err != nil {
		return nil, err
	}

	grpcFiles := r.GetFiles()
	files := make([]*handlers.File, 0, len(grpcFiles))
	for _, file := range grpcFiles {
		files = append(files, unmarshalFile(file))
	}

	return files, nil
}

// CreateStorage just create storage with specified storage name and permanent folder
func (c *GRPCFileServiceClient) CreateStorage(storage string, withPermanent bool) error {
	r, err := c.client.CreateStorage(context.Background(), &proto.CreateStorageRequest{
		Name:          storage,
		WithPermanent: withPermanent,
	})
	if err != nil {
		return err
	}

	if r.GetStatus() == proto.StorageStatus_AlreadyExist {
		return handlers.ErrAlreadyExist
	}

	return nil
}

// Remove remove file with fileName from storage
func (c *GRPCFileServiceClient) Remove(storage string, isPermanent bool, fileName string) error {
	_, err := c.client.RemoveFile(context.Background(), &proto.FileRequest{
		Storage:     storage,
		IsPermanent: isPermanent,
		FileName:    fileName,
	})

	return err
}

// Get download file from storage
func (c *GRPCFileServiceClient) Get(storage string, isPermanent bool, fileName string, w io.Writer) error {
	stream, err := c.client.GetFile(context.Background(), &proto.FileRequest{
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
func (c *GRPCFileServiceClient) Upload(storage string, isPermanent bool, fileName string, r io.Reader) (*handlers.File, error) {
	stream, err := c.client.UploadFile(context.Background())
	if err != nil {
		return nil, err
	}

	stream.Send(&proto.FileUploadRequest{
		FileChunk: &proto.FileUploadRequest_Metadata{
			Metadata: &proto.FileRequest{
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
			stream.Send(&proto.FileUploadRequest{
				FileChunk: &proto.FileUploadRequest_Content{
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

	err = stream.Send(&proto.FileUploadRequest{
		FileChunk: &proto.FileUploadRequest_End{
			End: true,
		},
	})

	if err != nil {
		return nil, err
	}

	file, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return unmarshalFile(file), nil
}

// IsStorageExists check specific storage for existanse
func (c *GRPCFileServiceClient) IsStorageExists(storage string) bool {
	r, err := c.client.IsStorageExists(context.Background(), &proto.IsStorageExistsRequest{
		Name: storage,
	})
	if err != nil {
		return false
	}

	return r.GetFlag()
}
