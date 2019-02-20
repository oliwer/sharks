package sharks

import (
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshPubKey struct {
	content []byte    // The public key as in `authorized_keys`
	date    time.Time // Date when this key was last indexed
}

// KeyCache is a map of SSH public keys indexed by their fingerprint.
type KeyCache struct {
	sync.Mutex
	keys  map[string]*sshPubKey
	debug bool
}

// NewKeyCache return a new KeyCache instance.
func NewKeyCache(debug bool) *KeyCache {
	return &KeyCache{
		keys:  make(map[string]*sshPubKey),
		debug: debug,
	}
}

// Clean up the cache by removing old entries. Return the number
// of elements deleted.
func (c *KeyCache) Clean(olderThan time.Duration) (removed int) {
	limit := time.Now().Add(-olderThan)

	for fgprint, key := range c.keys {
		if key.date.Before(limit) {
			c.Delete(fgprint)
			removed++
		}
	}

	return
}

// Delete an entry from the cache.
func (c *KeyCache) Delete(fingerprint string) {
	c.Lock()
	delete(c.keys, fingerprint)
	c.Unlock()
}

// Get an entry from the cache, and return a public key in the `AuthorizedKeys` format.
func (c *KeyCache) Get(fingerprint string) ([]byte, bool) {
	if sshpubkey, found := c.keys[fingerprint]; found {
		return sshpubkey.content, true
	}

	return nil, false
}

// Upsert a public key in the cache. Add it if it's new, or just update its date
// if it already exists. Return true if a new key is added.
func (c *KeyCache) Upsert(key []byte, name string) bool {
	c.Lock()
	defer c.Unlock()

	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		log.Printf("could not parse %s: %s\n", name, err)
		return false
	}

	fingerprint := ssh.FingerprintSHA256(publicKey)

	if sshpubkey, found := c.keys[fingerprint]; found {
		sshpubkey.date = time.Now()
		return false
	}

	c.keys[fingerprint] = &sshPubKey{
		content: ssh.MarshalAuthorizedKey(publicKey),
		date:    time.Now(),
	}
	if c.debug {
		log.Println("debug: added new key:", fingerprint)
	}
	return true
}
