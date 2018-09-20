package fs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

type Directory struct {
	Path string
}

func NewDirectory(path string) *Directory {
	return &Directory{
		Path: path,
	}
}

func (d *Directory) List() FileInfoList {
	osFiList, err := ioutil.ReadDir(d.Path)
	if err != nil {
		log.Println(err)
		return FileInfoList{}
	}

	fiList := make(FileInfoList, 0, len(osFiList))

	for _, osFi := range osFiList {
		fiList = append(fiList, FileInfo{osFi})
	}

	sort.Sort(fiList)

	return fiList
}

func (d *Directory) UniqueName(fileName string) string {
	ld := d.List()
	if !ld.Exist(fileName) {
		return fileName
	}

	ext := filepath.Ext(fileName)

	nameTemplate := fmt.Sprintf("%s%s%s", strings.TrimSuffix(fileName, ext), "_%d", ext)

	for count := 1; ; count++ {
		fileName = fmt.Sprintf(nameTemplate, count)
		if !ld.Exist(fileName) {
			break
		}
	}

	return fileName
}

func RunCleanWorker(dirPath string, protectedDir string, hour int, minute int) {
	now := time.Now()
	cleanTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, now.Second(), now.Nanosecond(), now.Location())
	go cleanDir(dirPath, protectedDir, cleanTime)
}

func cleanDir(dirPath string, protectedDir string, t time.Time) {
	if !path.IsAbs(dirPath) {
		dirPath, _ = filepath.Abs(dirPath)
	}

	tick := func() <-chan time.Time {
		now := time.Now()

		for t.Before(now) {
			t = t.Add(time.Hour * 24)
		}

		return time.After(t.Sub(now))
	}

	for {
		now := <-tick()
		storages, err := ioutil.ReadDir(dirPath)
		if err != nil {
			log.Println(err)
			return
		}

		for _, storage := range storages {
			if !storage.IsDir() {
				continue
			}

			sPath := path.Join(dirPath, storage.Name())

			log.Printf("time: %v; cleaning dir: %q\n", now, sPath)

			files, err := ioutil.ReadDir(sPath)
			if err != nil {
				log.Println(err)
				return
			}

			for _, file := range files {
				if file.IsDir() && file.Name() == protectedDir {
					continue
				}

				err = os.Remove(path.Join(sPath, file.Name()))
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}
