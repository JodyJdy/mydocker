package run

import (
	"containers"
	"fmt"
	"log"
	"os"
	"syscall"
)

func Run(tty bool, command string) {
	parent := containers.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Println(err)
	}
	// 等待parent进程执行完毕
	parent.Wait()
	os.Exit(-1)
}

// RunContainerInitProcess 执行容器内的进程
func RunContainerInitProcess(cmd string, args []string) error {
	log.Println("command is", cmd)
	//设置默认挂载参数 MS_NOEXEC本文将系统不允许运行其他程序 MS_NOEXEC 运行程序的时候
	// 不允许 set-user-id,set-group-id
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载 proc目录，使 容器有独立的proc目录
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
	return nil

}
