package cgroups

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
)

func FindCgroupMountPoint(subsystem string) string {
	// 获取挂载信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("关闭文件失败: %s\n", f.Name())
		}
	}(f)
	scanner := bufio.NewScanner(f)
	// 挂载信息格式如下
	// 34 25 0:30 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:16 - cgroup cgroup rw,memory
	// 最后的是 memory 类型，表示subsystem类型； 按照空格切分，第四个是 /sys/fs/cgroup/memory,表示顶层subsystem的hierarchy目录
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	// 获取cgroup顶层目录
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			// 创建子目录
			if err := os.MkdirAll(path.Join(cgroupRoot, cgroupPath), 0755); err == nil {
			} else {
				log.Printf("创建cgroup失败 %v\n", err)
				return "", err
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		log.Printf("cgroup 路径错误 %v\n", err)
		return "", err
	}
}
