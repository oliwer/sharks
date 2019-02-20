package sources

import (
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/AirVantage/sharks"
)

// ScanDirectory regularly reads `pubKeysDir` to populate the cache.
func ScanDirectory(cache *sharks.KeyCache, scanFrequency time.Duration, pubKeysDir string) {
	for {
		new := 0

		files, err := ioutil.ReadDir(pubKeysDir)
		if err != nil {
			log.Fatalln(err)
		}

		for _, file := range files {
			if file.Name()[0] == '.' {
				continue
			}

			bytes, err := ioutil.ReadFile(path.Join(pubKeysDir, file.Name()))
			if err != nil {
				log.Println(err)
				continue
			}

			if cache.Upsert(bytes, file.Name()) {
				new++
			}
		}

		if new > 0 {
			log.Printf("found %d new keys\n", new)
		}

		time.Sleep(scanFrequency)
	}
}
