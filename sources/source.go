package sources

import (
	"log"
	"time"
)

// Source of public keys.
type Source interface {
	// Initialize a Source and validate configuration.
	Init()
	// Scan for public keys, and return how many new ones were found.
	Scan() int
}

// KeepScanning a Source at a given frequency.
func KeepScanning(src Source, frequency time.Duration) {
	var new int

	for {
		new = src.Scan()

		if new > 0 {
			log.Printf("found %d new keys\n", new)
		}

		time.Sleep(frequency)
	}
}
