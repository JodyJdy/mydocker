package cgroups

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

func (c *CpuSubSystem) Name() string {
	return "cpu"
}

func (c *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			if err := os.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("设置 cgroup cpu share 失败 %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (c *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err == nil {
		if err := os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("设置 cgroup proc 失败 %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("获取 cgroup %s 失败: %v", cgroupPath, err)
	}
}

func (c *CpuSubSystem) Remove(cgroupPath string) error {
	// 删除cgroup相关的文件
	if subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err == nil {
		if _, err := os.Stat(subsysCgroupPath); err == nil {
			return os.Remove(subsysCgroupPath)
		}
		return err
	} else {
		return err
	}
}
