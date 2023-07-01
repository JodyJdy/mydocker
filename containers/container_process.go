package containers

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个父进程， 父进程的目的是
// 真正的执行cmd，并用cmd 对应的进程替换自身
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	// 调用mydocker的 init命令， 执行command
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	// 附加输入输出
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
