package run

import (
	"cgroups"
	"containers"
	"fmt"
	"io"
	"log"
	"networks"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func Run(tty bool, cmdArray []string, res *cgroups.ResourceConfig, volumes []string, containerName string, env []string, image string, portMapping []string, net string) {
	imageId := containers.ResolveImageId(image, false)
	command := containers.ResolveCmd(cmdArray, imageId, tty)
	// 提前获取容器id
	containerInfo := &containers.ContainerInfo{
		Id:          containers.ContainerId(),
		Command:     strings.Join(command.Cmds, " "),
		Status:      containers.Running,
		SetCgroup:   true,
		PortMapping: portMapping,
	}
	if containerName != "" {
		if containers.ResolveContainerId(containerName, true) != "" {
			fmt.Printf("容器名称重复 %s\n", containerName)
			return
		}
		containerInfo.Name = containerName
	}
	// 获取容器基础目录
	containerInfo.BaseUrl = fmt.Sprintf(containers.ContainerInfoLocation, containerInfo.Id)
	parent, writePipe := containers.NewParentProcess(containerInfo, tty, volumes, env, imageId)
	if parent == nil {
		log.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Printf("启动父进程失败:%v\n", err)
	}
	fmt.Printf("容器进程 pid: %d \n", parent.Process.Pid)
	// 记录容器信息
	containers.RecordContainerInfo(containerInfo, parent.Process.Pid)
	cgroups.ProcessCgroup(containerInfo.Id, parent.Process.Pid, res)

	if net != "" {
		networks.Init()
		processNetWork(net, command, containerInfo)
	}
	// 将命令写到管道里面
	containers.SendInitCommand(command, writePipe)
	if tty {
		// 等待parent进程执行完毕
		err := parent.Wait()
		if err != nil {
			fmt.Printf("等待父进程执行失败： %v\n", err)
			return
		}
		// 删除工作空间，卷的挂载点
		containers.DeleteWorkSpace(containerInfo)
		// 删除记录的容器信息
		containers.DeleteContainerInfo(containerInfo)
	}
	os.Exit(-1)
}
func processNetWork(net string, cmd *containers.CommandArray, info *containers.ContainerInfo) {
	// 什么都不做
	if net == networks.NONE {
		return
	}
	// 使用主机的网络
	if net == networks.HOST {
		cmd.Host = true
		return
	}
	// 和容器共享网络
	if strings.HasPrefix(net, networks.CONTAINER) {
		containerId := strings.Split(net, ":")[1]
		cmd.SharedNsContainer = containerId
		return
	}
	if err := networks.Connect(net, info); err != nil {
		fmt.Printf("连接网络失败%v\n", err)
		return
	}

}

// RunContainerInitProcess 执行容器内的进程
func RunContainerInitProcess() error {
	command := containers.ReadUserCommand()
	cmdArray := command.Cmds
	fmt.Printf("容器启动命令: %v\n", cmdArray)
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	// 设置当前进程的网络 当使用 -net host 和 -net container:id时
	setContainerNetNs(*command)
	// 初始化挂载信息
	containers.SetUpMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Printf("Exec loop path error %v\n", err)
		return err
	}
	//切换工作目录
	os.Chdir(command.WorkDir)
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
	return nil
}

// 设置容器的网络命名空间
func setContainerNetNs(cmd containers.CommandArray) {
	if cmd.SharedNsContainer != "" {
		infoId := containers.ResolveContainerId(cmd.SharedNsContainer, false)
		if infoId != "" {
			info, _ := containers.GetContainerInfo(infoId)
			networks.SetContainerNetNs(info.Pid)
		}
	} else if cmd.Host {
		// 使用1号进程的netns,也就是宿主机的网络
		networks.SetContainerNetNs(strconv.Itoa(1))
	}
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
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("读取日志文件: %s 失败  %v", logFileLocation, err)
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
