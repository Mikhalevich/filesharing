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

type Cleaner struct {
	Path         string
	ProtectedDir string
	finish       chan bool
}

func NewCleaner(path string, protectedDirPath string) *Cleaner {
	path, _ = filepath.Abs(path)

	return &Cleaner{
		Path:         path,
		ProtectedDir: protectedDirPath,
		finish:       make(chan bool),
	}
}

func (c *Cleaner) Run(hour int, minute int) {
	now := time.Now()
	cleanTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, now.Second(), now.Nanosecond(), now.Location())
	go c.clean(cleanTime)
}

func (c *Cleaner) Stop() {
	c.finish <- true
}

func (c *Cleaner) clean(t time.Time) {
	tick := func() <-chan time.Time {
		now := time.Now()

		for t.Before(now) {
			t = t.Add(time.Hour * 24)
		}

		return time.After(t.Sub(now))
	}

	for {
		var now time.Time
		select {
		case now = <-tick():
		case <-c.finish:
			log.Printf("Clean for %s is done\n", c.Path)
			return
		}

		log.Printf("time for cleaning: %v", now)

		storages, err := ioutil.ReadDir(c.Path)
		if err != nil {
			log.Println(err)
			return
		}

		for _, storage := range storages {
			if !storage.IsDir() {
				continue
			}

			sPath := path.Join(c.Path, storage.Name())

			log.Printf("cleaning dir: %q\n", sPath)

			files, err := ioutil.ReadDir(sPath)
			if err != nil {
				log.Println(err)
				return
			}

			for _, file := range files {
				if file.IsDir() && file.Name() == c.ProtectedDir {
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
