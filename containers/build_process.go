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
	info := &ContainerInfo{
		Id: containerId,
	}
	parent, writePipe := NewParentProcess(info, false, "", []string{}, image)
	if parent == nil {
		log.Println("New parent process error")
		return ""
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
	// 记录容器信息
	RecordContainerInfo(info, parent.Process.Pid)
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
	SendInitCommand(nil, writePipe)
	os.Exit(-1)
	return ""
}

// BuildFrom from时需要构建一个容器 使用
func BuildFrom2(image string) string {
	// 提前获取容器id
	//containerId := ContainerId()
	//parent, writePipe, _ := NewParentProcess(containerId, false, "", []string{}, image)
	//if parent == nil {
	//	log.Println("New parent process error")
	//	return ""
	//}
	//if err := parent.Start(); err != nil {
	//	fmt.Errorf("启动父进程失败:%v\n", err)
	//}
	//// 记录容器信息
	//RecordContainerInfo(containerId, parent.Process.Pid, []string{}, containerId, "", image)
	//// 将命令写到管道里面
	//SendInitCommand(nil, writePipe)
	//parent.Wait()
	return ""
}

// BuildRun  run时 需要进入容器执行命令
func BuildRun(containerId string, cmd []string) {
	ExecContainer(containerId, cmd)
}

// ReadUserCommand 从管道里面读取命令
func ReadUserCommand() *CommandArray {
	// 1个进程默认有三个文件描述符， 标准输入，标准输出，标准错误，文件描述符分别是 0,1,2
	// 当前读取的文件是第四个，文件描述符为3
	fmt.Println("读取管道文件")
	pipe := os.NewFile(uintptr(3), "pipe")
	command := new(CommandArray)
	return LoadCommand(command, pipe)
}

// SendInitCommand 将命令行信息写入到管道文件里面
func SendInitCommand(cmd *CommandArray, writePipe *os.File) {
	SaveCommand(cmd, writePipe)
}
