package main

import (
	"github.com/tberk-s/learning-url-shortener-with-go/src/apiserver"
)

func main() {
	apiserver.New(apiserver.WithServerEnv("development"))
}
