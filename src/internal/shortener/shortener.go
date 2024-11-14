package shortener

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

var urlMap = make(map[string]string)

func ShortenURL(originalURL string) string {
	// shorten the URL.

	hash := sha256.New()
	hash.Write([]byte(originalURL))
	fmt.Println(hash.Sum(nil))

	hashURL := hex.EncodeToString(hash.Sum(nil))
	shorterURL := hashURL[:6]

	urlMap[shorterURL] = originalURL

	return shorterURL
}

func GetOriginalURL(shortURL string) (string, bool) {
	originalURL, exists := urlMap[shortURL]
	return originalURL, exists
}
