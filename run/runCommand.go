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

func Run(tty bool, cmdArray []string, res *cgroups.ResourceConfig, volume string, containerName string) {
	// 提前获取容器id
	containerId := containers.ContainerId()
	parent, writePipe, containerBaseUrl := containers.NewParentProcess(containerId, tty, volume)
	if parent == nil {
		log.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
	// 记录容器信息
	containers.RecordContainerInfo(containerId, parent.Process.Pid, cmdArray, containerName, volume)
	// 创建cgroup manager
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Remove()
	//设置资源限制
	err := cgroupManager.Set(res)
	if err != nil {
		_ = fmt.Errorf("设置资源限制失败: %v \n", err)
		return
	}
	//将容器进程加入到cgroup中
	err = cgroupManager.Apply(parent.Process.Pid)
	if err != nil {
		_ = fmt.Errorf("添加容器进程到cgroup中失败: %v \n", err)
		return
	}
	// 将命令写到管道里面
	sendInitCommand(cmdArray, writePipe)
	if tty {
		// 等待parent进程执行完毕
		err = parent.Wait()
		if err != nil {
			_ = fmt.Errorf("等待父进程执行失败： %v", err)
			return
		}
		// 删除工作空间，卷的挂载点
		containers.DeleteWorkSpace(containerBaseUrl, volume)
		// 删除记录的容器信息
		containers.DeleteContainerInfo(containerId)
	}
	os.Exit(-1)
}

// 将命令行信息写入到管道文件里面
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	fmt.Printf("command all is %s\n", command)
	_, err := writePipe.WriteString(command)
	if err != nil {
		return
	}
	err = writePipe.Close()
	if err != nil {
		return
	}
}

// RunContainerInitProcess 执行容器内的进程
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	fmt.Printf("命令行是:%v\n", cmdArray)
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	// 初始化挂载信息
	containers.SetUpMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Printf("Exec loop path error %v\n", err)
		return err
	}
	fmt.Printf("Find path %s\n", path)
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
	return nil
}

// 从管道里面读取命令
func readUserCommand() []string {
	// 1个进程默认有三个文件描述符， 标准输入，标准输出，标准错误，文件描述符分别是 0,1,2
	// 当前读取的文件是第四个，文件描述符为3
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := io.ReadAll(pipe)
	if err != nil {
		fmt.Printf("init read pipe error %v\n", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// Ps 列出所有进程
func Ps() {
	containers.ListContainerInfo()
}

// Log 显示container的日志，先按照容器id打开
func Log(containerId string) {
	dirURL := fmt.Sprintf(containers.DefaultInfoLocation, containerId)
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
func Exec(containerId string, cmdArray []string) {
	containers.ExecContainer(containerId, cmdArray)
}

// Stop 停止容器
func Stop(containerId string) {
	containers.StopContainer(containerId)
}

// Remove 删除容器
func Remove(containerId string) {
	containers.RemoveContainer(containerId)
}
