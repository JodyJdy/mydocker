package containers

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// ContainerId 生成容器id
func ContainerId() string {
	return randStringBytes(10)
}

// VolumeId 生成默认卷id
func VolumeId() string {
	return randStringBytes(5)
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
func Copy(from, to string) error {
	f, e := os.Stat(from)
	if e != nil {
		return e
	}
	if f.IsDir() {
		//from是文件夹，那么定义to也是文件夹g
		if list, e := os.ReadDir(from); e == nil {
			for _, item := range list {
				if e = Copy(filepath.Join(from, item.Name()), filepath.Join(to, item.Name())); e != nil {
					return e
				}
			}
		}
	} else {
		//from是文件，那么创建to的文件夹g
		p := filepath.Dir(to)
		if _, e = os.Stat(p); e != nil {
			if e = os.MkdirAll(p, 0777); e != nil {
				return e
			}
		}
		//读取源文件g
		file, e := os.Open(from)
		if e != nil {
			return e
		}
		defer file.Close()
		bufReader := bufio.NewReader(file)
		// 创建一个文件用于保存
		out, e := os.Create(to)
		if e != nil {
			return e
		}
		defer out.Close()
		// 然后将文件流和文件流对接起来g
		_, e = io.Copy(out, bufReader)
	}
	return e
}
