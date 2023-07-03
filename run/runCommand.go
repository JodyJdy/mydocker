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

func Run(tty bool, cmdArray []string, res *cgroups.ResourceConfig) {
	parent, writePipe := containers.NewParentProcess(tty)
	if parent == nil {
		log.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Println(err)
	}
	// 创建cgroup manager
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Remove()
	//设置资源限制
	cgroupManager.Set(res)
	//将容器进程加入到cgroup中
	cgroupManager.Apply(parent.Process.Pid)
	// 将命令写到管道里面
	sendInitCommand(cmdArray, writePipe)
	// 等待parent进程执行完毕
	parent.Wait()
	os.Exit(-1)
}

// 将命令行信息写入到管道文件里面
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	fmt.Println("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

// RunContainerInitProcess 执行容器内的进程
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error, cmdArray is nil")
	}
	//设置默认挂载参数 MS_NOEXEC本文将系统不允许运行其他程序 MS_NOEXEC 运行程序的时候
	// 不允许 set-user-id,set-group-id
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载 proc目录，使 容器有独立的proc目录
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Println("Exec loop path error %v", err)
		return err
	}
	fmt.Println("Find path %s", path)
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
		fmt.Println("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
