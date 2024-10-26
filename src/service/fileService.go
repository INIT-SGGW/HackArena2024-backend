package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type FileService struct {
	StoragePath                  string
	pathToAllSolutionTempStorage string
	logger                       *zap.Logger
}

func NewFileService(logger *zap.Logger, pathToStorage, pathToAllSolutionTempStorage string) *FileService {
	return &FileService{
		logger:                       logger,
		StoragePath:                  pathToStorage,
		pathToAllSolutionTempStorage: pathToAllSolutionTempStorage,
	}

}

// TODO: move logic with disk file saving downloading operations
func (fs FileService) CopyFile(src, dst string) error {
	defer fs.logger.Sync()

	fs.logger.Info("Get stat of source file",
		zap.String("srcPath", src))
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		fs.logger.Error("Get stat of source file",
			zap.String("srcPath", src),
			zap.Error(err))
		wraped := fmt.Errorf("[CopyFile] error in geting stats %w", err)

		return wraped
	}
	if !sourceFileStat.Mode().IsRegular() {
		fs.logger.Error("The source file is not regular",
			zap.String("srcPath", src),
			zap.Error(err))
		wraped := fmt.Errorf("[CopyFile] source file is not regular %w", err)

		return wraped
	}
	source, err := os.Open(src)
	if err != nil {
		fs.logger.Error("The source file open error",
			zap.String("srcPath", src),
			zap.Error(err))
		wraped := fmt.Errorf("[CopyFile] source file open error %w", err)

		return wraped
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		fs.logger.Error("error creating destintion file",
			zap.String("dstPath", dst),
			zap.Error(err))
		wraped := fmt.Errorf("[CopyFile] error creating destintion file %w", err)

		return wraped
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		fs.logger.Error("error while copy the files",
			zap.String("srcPath", src),
			zap.String("dstPath", dst),
			zap.Error(err))
	} else {
		fs.logger.Info("The file was sucesfully copied",
			zap.String("srcPath", src),
			zap.String("dstPath", dst))
	}
	return err
}

func (fs FileService) ZipArchive(dir, archivePath string) error {
	defer fs.logger.Sync()

	f, err := os.Create(archivePath)
	if err != nil {
		fs.logger.Error("error in creating zip archive file",
			zap.Error(err))
		wraped := fmt.Errorf("[ZipArchive] error in creating zip archive file %w", err)

		return wraped
	}
	defer f.Close()
	fs.logger.Info("Creating file for archiving")

	archiveWriter := zip.NewWriter(f)
	defer archiveWriter.Close()

	// TODO Add more insigts in logs
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			wraped := fmt.Errorf("[ZipArchive] error in file walking %w", err)

			return wraped
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			wraped := fmt.Errorf("[ZipArchive] error creating local file header %w", err)

			return wraped
		}

		// set compression
		header.Method = zip.Deflate

		header.Name, err = filepath.Rel(filepath.Dir(dir), path)
		if err != nil {
			wraped := fmt.Errorf("[ZipArchive] error setting new path%w", err)

			return wraped
		}
		if info.IsDir() {
			header.Name += "/"
		}

		headerWriter, err := archiveWriter.CreateHeader(header)
		if err != nil {
			wraped := fmt.Errorf("[ZipArchive] error saving content to the file %w", err)

			return wraped
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})

}
