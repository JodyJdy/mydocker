package run

import (
	"cgroups"
	"containers"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Run(tty bool, cmdArray []string, res *cgroups.ResourceConfig, volume string, containerName string, env []string, image string) {
	command := containers.ResolveCmd(cmdArray, image, tty)
	fmt.Println("要执行的命令")
	fmt.Println(command)
	// 提前获取容器id
	containerInfo := &containers.ContainerInfo{
		Id:        containers.ContainerId(),
		Command:   strings.Join(command.Cmds, " "),
		Status:    containers.Running,
		SetCgroup: true,
	}
	if containerName != "" {
		if containers.ResolveContainerId(containerName, true) != "" {
			fmt.Printf("镜像名称重复 %s\n", containerName)
			return
		}
		containerInfo.Name = containerName
	}
	// 获取容器基础目录
	containerInfo.BaseUrl = fmt.Sprintf(containers.ContainerInfoLocation, containerInfo.Id)
	parent, writePipe := containers.NewParentProcess(containerInfo, tty, volume, env, image)
	if parent == nil {
		log.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Printf("启动父进程失败:%v\n", err)
	}
	fmt.Printf("容器id: %d", parent.Process.Pid)
	// 记录容器信息
	containers.RecordContainerInfo(containerInfo, parent.Process.Pid)
	containers.ProcessCgroup(containerInfo, parent.Process.Pid, res)
	// 将命令写到管道里面
	containers.SendInitCommand(command, writePipe)
	if tty {
		// 等待parent进程执行完毕
		err := parent.Wait()
		if err != nil {
			fmt.Printf("等待父进程执行失败： %v\n", err)
		}
		// 删除工作空间，卷的挂载点
		containers.DeleteWorkSpace(containerInfo)
		// 删除记录的容器信息
		containers.DeleteContainerInfo(containerInfo)
	}
	os.Exit(-1)
}

// RunContainerInitProcess 执行容器内的进程
func RunContainerInitProcess() error {
	command := containers.ReadUserCommand()
	cmdArray := command.Cmds
	fmt.Printf("命令行是:%v\n", cmdArray)
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	// 初始化挂载信息
	containers.SetUpMount()
	//切换工作目录
	os.Chdir(command.WorkDir)
	path, err := exec.LookPath(cmdArray[0])
	fmt.Println("Find path %s", path)
	if err != nil {
		fmt.Printf("Exec loop path error %v\n", err)
		return err
	}
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
	return nil
}

// Ps 列出所有进程
func Ps() {
	containers.ListContainerInfo()
}

// Log 显示container的日志，先按照容器id打开
func Log(idOrName string) {
	containerId := containers.ResolveContainerId(idOrName, false)
	if containerId == "" {
		fmt.Printf("无法根据提供的容器标识定位到容器\n")
		return
	}
	dirURL := fmt.Sprintf(containers.ContainerInfoLocation, containerId)
	logFileLocation := dirURL + containers.ContainerLogName
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		fmt.Errorf("log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Errorf("log container read file %s error %v", logFileLocation, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}

// Exec 进入容器
func Exec(idOrName string, cmdArray []string) {
	containerId := containers.ResolveContainerId(idOrName, false)
	if containerId == "" {
		fmt.Printf("无法根据提供的容器标识定位到容器\n")
		return
	}
	containers.ExecContainer(containerId, cmdArray)
}

// Stop 停止容器
func Stop(idOrName string) {
	containerId := containers.ResolveContainerId(idOrName, false)
	if containerId == "" {
		fmt.Printf("无法根据提供的容器标识定位到容器\n")
		return
	}
	containers.StopContainer(containerId)
}

// Remove 删除容器
func Remove(idOrName string) {
	containerId := containers.ResolveContainerId(idOrName, false)
	if containerId == "" {
		fmt.Printf("无法根据提供的容器标识定位到容器\n")
		return
	}
	containers.RemoveContainer(containerId)
}
