package fileInfo

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
	Path string
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

func ListDir(dirPath string) FileInfoList {
	osFiList, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Println(err.Error)

		return FileInfoList{}
	}

	fiList := make(FileInfoList, 0, len(osFiList))

	for _, osFi := range osFiList {
		fi := FileInfo{osFi, path.Join(dirPath, osFi.Name())}
		fiList = append(fiList, fi)
	}

	sort.Sort(fiList)

	return fiList
}

func CleanDir(dirPath string, t time.Time) {
	if !path.IsAbs(dirPath) {
		dirPath, _ = filepath.Abs(dirPath)
	}

	tick := func() <-chan time.Time {
		now := time.Now()

		if t.Before(now) {
			t = t.Add(time.Hour * 24)
		}

		return time.Tick(t.Sub(now))
	}

	for {
		c := tick()

		for now := range c {
			if err := os.RemoveAll(dirPath); err != nil {
				log.Println(err.Error())

				return
			}

			log.Printf("time: %v; cleaning dir: %q\n", now, dirPath)

			if err := os.Mkdir(dirPath, 0777); err != nil {
				log.Println(err.Error())

				return
			}

			break
		}
	}
}

func UniqueName(fileName string, dir string) string {
	if ld := ListDir(dir); ld.Exist(fileName) {
		ext := filepath.Ext(fileName)

		fileNameTpl := strings.TrimSuffix(fileName, ext) + "_%d" + ext

		count := 1

		var f func() string
		f = func() string {
			fileName = fmt.Sprintf(fileNameTpl, count)
			if ld.Exist(fileName) {
				count++
				return f()
			}

			return fileName
		}

		return f()
	}

	return fileName
}
