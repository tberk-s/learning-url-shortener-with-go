package shortener

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func ShortenURL(originalURL string) string {
	// shorten the URL.

	hash := sha256.New()
	hash.Write([]byte(originalURL))
	fmt.Println(hash.Sum(nil))

	hashURL := hex.EncodeToString(hash.Sum(nil))
	shorterURL := hashURL[:6]

	return shorterURL
}
