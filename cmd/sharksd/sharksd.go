package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/AirVantage/sharks"
	"github.com/AirVantage/sharks/sources"
)

var (
	debug         = flag.Bool("debug", false, "activate debug mode")
	listenAddr    = flag.String("listen", ":8080", "Listening address of the HTTP server")
	logFlags      = flag.Int("logflags", 0, "Go's logger flags")
	pubKeysDir    = flag.String("keysdir", "sshkeys", "Directory where the public SSH keys are stored")
	s3bucket      = flag.String("s3bucket", "", "S3 Bucket and Prefix where to look for SSH public keys")
	s3region      = flag.String("s3region", "eu-west-1", "AWS Region for S3")
	scanFrequency = flag.Duration("freq", 1*time.Minute, "Frequency to scan the keys directory")

	cache *sharks.KeyCache
)

// Remove expired ssh keys from cache regularly.
func cacheCleaner() {
	for {
		removed := cache.Clean(*scanFrequency * 2)

		if removed > 0 {
			log.Printf("removed %d expired keys\n", removed)
		}

		time.Sleep(*scanFrequency * 2)
	}
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func lookupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	fgprint := r.URL.Query().Get("fingerprint")

	if *debug {
		log.Println("debug: /lookup fingerprint:", fgprint)
	}

	if len(fgprint) == 0 || len(fgprint) > 100 {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}

	key, found := cache.Get(fgprint)

	if found {
		w.Write(key)
	} else {
		http.NotFound(w, r)
	}

	if *debug {
		log.Println("debug: /lookup status:", found)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(*logFlags)
	cache = sharks.NewKeyCache(*debug)

	go cacheCleaner()

	// Select the source of SSH keys
	if *s3bucket != "" {
		go sources.ScanS3Bucket(cache, *scanFrequency, *s3bucket, *s3region)
	} else {
		go sources.ScanDirectory(cache, *scanFrequency, *pubKeysDir)
	}

	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/lookup", lookupHandler)
	log.Println(http.ListenAndServe(*listenAddr, nil))
}
