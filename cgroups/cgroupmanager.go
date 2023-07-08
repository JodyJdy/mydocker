package cgroups

import "log"

// cgroup根路径
var Roout_Cgroup_Path = "mydocker-cgroup/"

type CgroupManager struct {
	// cgroup在hierarchy中的路径 相当于创建的cgroup目录相对于root cgroup目录的路径
	Path string
	// 资源配置
	Resource *ResourceConfig
}

// NewCgroupManager 创建一个新的Manager对象
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply 进程添加到cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range SubsystemIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set 设置cgroup资源限制
func (c *CgroupManager) Set(res *ResourceConfig) error {
	for _, subSysIns := range SubsystemIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

// Remove 释放cgroup
func (c *CgroupManager) Remove() {
	for _, subSysIns := range SubsystemIns {
		err := subSysIns.Remove(c.Path)
		if err != nil {
			log.Fatalf("remove cgroup fail %v", err)
			return
		}
	}
}
