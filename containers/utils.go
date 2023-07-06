package containers

import (
	"math/rand"
	"os"
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

// ImageId 生成镜像id, 生成15位的镜像id；与 容器id作区分
func ImageId() string {
	return randStringBytes(15)
}

// GetBaseImageId 最基础的镜像id,为和其他镜像区分，名称不使用数字
func GetBaseImageId() string {
	return "base"
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
