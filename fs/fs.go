package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type ByteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

var (
	PermanentDir string

	ErrNotExists = errors.New("Not exists")
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

type FileInfo struct {
	os.FileInfo
}

func (fi *FileInfo) Size() string {
	return ByteSize(fi.FileInfo.Size()).String()
}

type FileInfoList []FileInfo

func (fil FileInfoList) Len() int {
	return len(fil)
}

func (fil FileInfoList) Swap(i, j int) {
	fil[i], fil[j] = fil[j], fil[i]
}

func (fil FileInfoList) Less(i, j int) bool {
	if PermanentDir != "" {
		if fil[i].IsDir() && fil[i].Name() == PermanentDir {
			return true
		}

		if fil[j].IsDir() && fil[j].Name() == PermanentDir {
			return false
		}
	}

	return fil[i].ModTime().After(fil[j].ModTime())
}

func (fil FileInfoList) Exist(name string) bool {
	for _, fi := range fil {
		if fi.Name() == name {
			return true
		}
	}

	return false
}

type FileStorage struct {
}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) Files(path string) FileInfoList {
	return newDirectory(path).List()
}

func (fs *FileStorage) Store(dir string, fileName string, data io.Reader) (string, error) {
	uniqueName := newDirectory(dir).UniqueName(fileName)
	f, err := os.Create(path.Join(dir, uniqueName))
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return "", err
	}

	return uniqueName, nil
}

func (fs *FileStorage) Move(filePath string, dir string, fileName string) error {
	uniqueName := newDirectory(dir).UniqueName(fileName)
	return os.Rename(filePath, path.Join(dir, uniqueName))
}

func (fs *FileStorage) Remove(dir string, fileName string) error {
	files := newDirectory(dir).List()
	if !files.Exist(fileName) {
		return ErrNotExists
	}

	return os.Remove(path.Join(dir, fileName))
}
