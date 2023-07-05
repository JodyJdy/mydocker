package containers

import (
	"math/rand"
	"time"
)

// ContainerId 生成容器id
func ContainerId() string {
	return randStringBytes(10)
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
