package misc

import (
    "math/rand"
    "time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = *rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomFilename(length int) string {
    s := make([]byte, length)
    for i := range s {
        s[i] = chars[seededRand.Intn(len(chars))]
    }
    return string(s)
}