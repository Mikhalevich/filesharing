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
	rootPath string
}

func NewFileStorage(root string) *FileStorage {
	return &FileStorage{
		rootPath: root,
	}
}

func (fs *FileStorage) Root() string {
	return fs.rootPath
}

func (fs *FileStorage) Join(p string) string {
	return path.Join(fs.rootPath, p)
}

func (fs *FileStorage) IsExists(p string) bool {
	_, err := os.Stat(fs.Join(p))
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func (fs *FileStorage) Mkdir(dir string) error {
	return os.Mkdir(fs.Join(dir), os.ModePerm)
}

func (fs *FileStorage) Files(p string) FileInfoList {
	return newDirectory(fs.Join(p)).List()
}

func (fs *FileStorage) Store(dir string, fileName string, data io.Reader) (string, error) {
	dirPath := fs.Join(dir)
	uniqueName := newDirectory(dirPath).UniqueName(fileName)
	f, err := os.Create(path.Join(dirPath, uniqueName))
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
	dirPath := fs.Join(dir)
	uniqueName := newDirectory(dirPath).UniqueName(fileName)
	return os.Rename(filePath, path.Join(dirPath, uniqueName))
}

func (fs *FileStorage) Remove(dir string, fileName string) error {
	dirPath := fs.Join(dir)
	files := newDirectory(dirPath).List()
	if !files.Exist(fileName) {
		return ErrNotExists
	}

	return os.Remove(path.Join(dirPath, fileName))
}
