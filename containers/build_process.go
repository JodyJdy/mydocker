package containers

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
)

func BuildFrom(image string) *ContainerInfo {
	imageId := resolveImageId(image)
	if imageId == "" {
		fmt.Printf("镜像不存在:%s \n", image)
		return nil
	}
	// 空命令
	command := &CommandArray{
		Cmds:    []string{"sh"},
		WorkDir: "/",
	}
	// 提前获取容器id
	containerId := ContainerId()
	fmt.Printf("容器id:%s", containerId)
	info := &ContainerInfo{
		Id:     containerId,
		Status: Stop,
	}
	// 获取容器基础目录
	info.BaseUrl = fmt.Sprintf(ContainerInfoLocation, info.Id)
	parent, writePipe := NewParentProcess(info, false, "", []string{}, imageId)
	if parent == nil {
		log.Println("New parent process error")
		return nil
	}
	if err := parent.Start(); err != nil {
		fmt.Printf("启动父进程失败:%v\n", err)
	}
	RecordContainerInfo(info, parent.Process.Pid)
	// 将命令写到管道里面
	SendInitCommand(command, writePipe)
	return info
}
func BuildRun(d *DockerFile, command *CommandArray) {
	parent, writePipe := RunParentProcess(d.Info, d.Env, d.WorkDir)
	if parent == nil {
		log.Println("New run parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
	RecordContainerInfo(d.Info, parent.Process.Pid)
	// 将命令写到管道里面
	SendInitCommand(command, writePipe)
	// 存在有多个Run的情况，需要等待上一个执行完毕
	parent.Wait()
}
func RunParentProcess(info *ContainerInfo, env []string, workDir string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		_ = fmt.Errorf("new pipe error %v", err)
		return nil, nil
	}
	// 调用mydocker的 init命令， 执行command
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	// 用于读取 管道中的命令
	cmd.ExtraFiles = []*os.File{readPipe}
	// 写入日志
	logFilePath := info.BaseUrl + ContainerLogName
	logFile, err := os.Create(logFilePath)
	fmt.Printf("写入日志路径:%s\n", logFilePath)
	// 将进程的输出重定向到logFile中，访问这个文件，就能读取到日志
	cmd.Stdout = logFile
	// 设置环境变量
	cmd.Env = append(os.Environ(), env...)
	cmd.Dir = path.Join(info.BaseUrl, "merged", workDir)
	return cmd, writePipe
}

func resolveImageId(from string) string {
	infoList := GetImageInfoList()
	for _, info := range infoList {
		imageName := info.Name
		if info.Version != "" {
			imageName += ":" + info.Version
		}
		if imageName == from {
			return info.Id
		}
	}
	return ""
}
