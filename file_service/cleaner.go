package fs

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

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
