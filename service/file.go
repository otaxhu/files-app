package service

import (
	"context"
	"io"

	"github.com/oklog/ulid/v2"
	"github.com/otaxhu/files-app/dto"
	"github.com/otaxhu/files-app/models"
	"github.com/otaxhu/files-app/repository"
)

type FileService struct {
	fileRepo repository.FileRepository
}

func NewFileService(fileRepo repository.FileRepository) *FileService {
	return &FileService{
		fileRepo: fileRepo,
	}
}

func (f *FileService) SaveFile(ctx context.Context, file dto.SaveFile) (id string, err error) {
	if file.Len <= 0 {
		err = ErrInvalidFile
		return
	}

	id = ulid.Make().String()

	err = f.fileRepo.SaveFile(ctx, models.File{
		Id:       id,
		Filename: file.Filename,
		Reader:   io.NopCloser(file.Reader),
		Len:      file.Len,
	})
	return
}

func (f *FileService) GetFile(ctx context.Context, id string) (*dto.GetFile, error) {
	file, err := f.fileRepo.GetFile(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.GetFile{
		Filename: file.Filename,
		Reader:   file.Reader,
		Len:      file.Len,
	}, nil
}

func (f *FileService) GetFileInfo(ctx context.Context, id string) (*dto.GetFileInfo, error) {
	file, err := f.fileRepo.GetFileInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.GetFileInfo{
		Filename: file.Filename,
		Len:      file.Len,
	}, nil
}

func (f *FileService) DeleteFile(ctx context.Context, id string) error {
	err := f.fileRepo.DeleteFile(ctx, id)
	return err
}
