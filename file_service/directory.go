package fs

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

type directory struct {
	Path string
}

func newDirectory(path string) *directory {
	return &directory{
		Path: path,
	}
}

func (d *directory) List() FileInfoList {
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

func (d *directory) UniqueName(fileName string) string {
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
