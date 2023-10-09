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
		// 使用宿主机的网络
		Host: true,
	}
	// 提前获取容器id
	containerId := ContainerId()
	info := &ContainerInfo{
		Id:     containerId,
		Status: Stop,
	}
	// 获取容器基础目录
	info.BaseUrl = fmt.Sprintf(ContainerInfoLocation, info.Id)
	parent, writePipe := NewParentProcess(info, false, []string{}, []string{}, imageId)
	if parent == nil {
		log.Println("启动父进程失败")
		return nil
	}
	//初始化镜像构建使用的域名解析文件
	CopyFile("/etc/resolv.conf", path.Join(parent.Dir, "/etc/resolv.conf"))
	parent.Stdout = os.Stdout
	parent.Stderr = os.Stderr
	if err := parent.Start(); err != nil {
		fmt.Printf("启动父进程失败:%v\n", err)
	}
	RecordContainerInfo(info, parent.Process.Pid)
	// 将命令写到管道里面
	SendInitCommand(command, writePipe)
	return info
}
func BuildRun(d *DockerFile, command *CommandArray) {
	command.Host = true
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
	//获取构建过程中的输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 设置环境变量
	cmd.Env = append(os.Environ(), env...)
	//这个目录是容器的root目录，不拼接 workdir
	cmd.Dir = path.Join(info.BaseUrl, "merged")
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
