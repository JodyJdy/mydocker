package containers

import (
	"cgroups"
	"fmt"
	"log"
	"os"
)

func BuildFrom(image string) string {
	// 提前获取容器id
	containerId := ContainerId()
	parent, writePipe, _ := NewParentProcess(containerId, false, "", []string{}, image)
	if parent == nil {
		log.Println("New parent process error")
		return ""
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
	// 记录容器信息
	RecordContainerInfo(containerId, parent.Process.Pid, []string{}, containerId, "", image)
	// 创建cgroup manager
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Remove()
	//设置资源限制
	err := cgroupManager.Set(&cgroups.ResourceConfig{
		MemoryLimit: "",
		CpuSet:      "",
		CpuShare:    "",
	})
	if err != nil {
		_ = fmt.Errorf("设置资源限制失败: %v \n", err)
		return ""
	}
	//将容器进程加入到cgroup中
	err = cgroupManager.Apply(parent.Process.Pid)
	if err != nil {
		_ = fmt.Errorf("添加容器进程到cgroup中失败: %v \n", err)
		return ""
	}
	// 将命令写到管道里面
	sendInitCommand("sleep 99999", writePipe)
	os.Exit(-1)
	return ""
}

// BuildFrom from时需要构建一个容器 使用
func BuildFrom2(image string) string {
	// 提前获取容器id
	containerId := ContainerId()
	parent, writePipe, _ := NewParentProcess(containerId, false, "", []string{}, image)
	if parent == nil {
		log.Println("New parent process error")
		return ""
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
	// 记录容器信息
	RecordContainerInfo(containerId, parent.Process.Pid, []string{}, containerId, "", image)
	// 将命令写到管道里面
	sendInitCommand("sh", writePipe)
	parent.Wait()
	return containerId
}

// BuildRun  run时 需要进入容器执行命令
func BuildRun(containerId string, cmd []string) {
	ExecContainer(containerId, cmd)
}

func sendInitCommand(cmd string, writePipe *os.File) {
	fmt.Printf("command is %s\n", cmd)
	_, err := writePipe.WriteString(cmd)
	if err != nil {
		return
	}
	err = writePipe.Close()
	if err != nil {
		return
	}
}
