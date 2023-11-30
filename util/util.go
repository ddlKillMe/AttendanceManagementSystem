package util

import (
	"math/rand"
	"time"
)

func RandomName(n int) string {
	letters := []byte("qwertyuiopQWERTYUIOPasdfghjklASDFGHJKLzxcvbnmZXCVBNM")
	res := make([]byte, n)

	rand.NewSource(time.Now().Unix())
	for i := range res {
		res[i] = letters[rand.Intn(len(letters))]
	}

	return string(res)
}
