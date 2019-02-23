package sources

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/AirVantage/sharks"
)

type Filesystem struct {
	Cache *sharks.KeyCache
	Path  string
}

func (fs *Filesystem) Init() {
	dir, err := os.Open(fs.Path)
	if err != nil {
		log.Fatalln(err)
	}
	info, _ := dir.Stat()
	if !info.IsDir() {
		log.Fatalln(fs.Path, "must be a directory")
	}
}

func (fs *Filesystem) Scan() int {
	new := 0

	files, err := ioutil.ReadDir(fs.Path)
	if err != nil {
		log.Println(err)
		return 0
	}

	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}

		bytes, err := ioutil.ReadFile(path.Join(fs.Path, file.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		if fs.Cache.Upsert(bytes, file.Name()) {
			new++
		}
	}

	return new
}
