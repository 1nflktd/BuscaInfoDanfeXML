package main

import (
    "os"
    "io/ioutil"
	"math/rand"
    "time"
)

func generateUniqueString(n int) string {
    rand.Seed(time.Now().UnixNano())

    const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }

    return string(b)
}

func getFile(src string) ([]byte, error) {
	return ioutil.ReadFile(src)
}

func isDirectory(src string) (bool, error) {
	fileInfo, err := os.Stat(src)
	if err == nil {
	    return fileInfo.IsDir(), err
	}
	return false, err
}
