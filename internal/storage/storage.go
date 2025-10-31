package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const (
	MaxFileSize = 500 * 1024 * 1024 // 500MB - large enough for practical datasets
)

type FileStore struct {
	baseDir string
}

func NewFileStore() (*FileStore, error) {
	baseDir := getStorageDir()

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStore{baseDir: baseDir}, nil
}

func getStorageDir() string {
	var baseDir string

	switch runtime.GOOS {
	case "darwin", "linux":
		baseDir = "/opt/iot-data-sandbox"
	case "windows":
		baseDir = filepath.Join(os.Getenv("ProgramData"), "iot-data-sandbox")
	default:
		// Fallback to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			baseDir = "app-files"
		} else {
			baseDir = filepath.Join(cwd, "app-files")
		}
	}

	// If we can't create the system directory, fall back to local directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		cwd, _ := os.Getwd()
		baseDir = filepath.Join(cwd, "app-files")
	}

	return baseDir
}

func (fs *FileStore) SaveFile(filename string, reader io.Reader, maxSize int64) (string, error) {
	if maxSize <= 0 {
		maxSize = MaxFileSize
	}

	// Create unique filename to avoid collisions
	destPath := filepath.Join(fs.baseDir, filename)

	// Check if file already exists and create unique name if needed
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(filename)
		base := filename[:len(filename)-len(ext)]
		for i := 1; ; i++ {
			newName := fmt.Sprintf("%s_%d%s", base, i, ext)
			destPath = filepath.Join(fs.baseDir, newName)
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				filename = newName
				break
			}
		}
	}

	file, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Limit read size to prevent excessive memory usage
	limitedReader := io.LimitReader(reader, maxSize)
	written, err := io.Copy(file, limitedReader)
	if err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if written == maxSize {
		os.Remove(destPath)
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	return filename, nil
}

func (fs *FileStore) GetFilePath(filename string) string {
	return filepath.Join(fs.baseDir, filename)
}

func (fs *FileStore) DeleteFile(filename string) error {
	path := filepath.Join(fs.baseDir, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (fs *FileStore) FileExists(filename string) bool {
	path := filepath.Join(fs.baseDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

func (fs *FileStore) GetBaseDir() string {
	return fs.baseDir
}
