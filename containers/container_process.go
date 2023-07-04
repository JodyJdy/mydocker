package containers

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个父进程， 父进程的目的是
// 真正的执行cmd，并用cmd 对应的进程替换自身
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	// 附加输入输出
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	// 用于读取 管道中的命令
	cmd.ExtraFiles = []*os.File{readPipe}
	// 容器工作目录， @todo 应该每个容器一个目录
	containerBaseUrl := "/root/"
	// 工作目录，为 overlay文件系统中的 merge目录
	// 容器进程，会以merged目录作为根目录运行
	cmd.Dir = NewWorkSpace(containerBaseUrl)
	return cmd, writePipe
}

// NewPipe 创建管道对象
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
