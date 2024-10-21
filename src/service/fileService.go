package service

import "go.uber.org/zap"

type FileService struct {
	StoragePath string
	logger      *zap.Logger
}

func NewFileService(logger *zap.Logger, pathToStorage string) *FileService {
	return &FileService{
		logger:      logger,
		StoragePath: pathToStorage,
	}

}

// TODO: move logic with disk file saving downloading operations
