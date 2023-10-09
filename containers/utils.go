package containers

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
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

// CopyFile 单文件拷贝
func CopyFile(fromFile, toFile string) {
	toDir := path.Dir(toFile)
	_, err := os.Stat(path.Dir(toFile))
	// 如果不存在进行创建
	if err != nil {
		err := os.MkdirAll(toDir, 0622)
		if err != nil {
			log.Printf("创建文件夹失败:%s, 原因: %v\n", toDir, err)
			return
		}
	}
	copyCmd := fmt.Sprintf("cp   %s  %s", fromFile, toFile)
	cmd := exec.Command("sh", "-c", copyCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("拷贝文件错误:%s, %v\n", out, err)
	}
}

// Copy 拷贝到目录
func Copy(from, to string) {
	info, err := os.Stat(to)
	// 如果不存在进行创建
	if err != nil {
		err := os.MkdirAll(to, 0622)
		if err != nil {
			log.Printf("创建文件夹失败:%s, 原因: %v\n", to, err)
			return
		}
	} else {
		if !info.IsDir() {
			log.Printf("%s不是一个目录\n", to)
			return
		}
	}
	copyCmd := fmt.Sprintf("cp -R  %s  %s", from, to)
	cmd := exec.Command("sh", "-c", copyCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("拷贝文件错误:%s, %v\n", out, err)
	}
}

// UnTar 解压文件
func UnTar(from, to string) {
	info, err := os.Stat(to)
	// 如果不存在进行创建
	if err != nil {
		err := os.MkdirAll(to, 0622)
		if err != nil {
			log.Printf("创建文件夹失败:%s, 原因: %v\n", to, err)
			return
		}
	} else {
		if !info.IsDir() {
			log.Printf("%s不是一个目录\n", to)
			return
		}
	}

	cmd := exec.Command("tar", "-xvf", from, "-C", to)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("解压文件错误:%s, %v\n", out, err)
	}
}
