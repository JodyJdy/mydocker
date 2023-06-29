package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	testNet()
}

// 测试 uts 命令空间
func testUts() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//创建 uts ns
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// 测试 ipc命名空间
func testIpc() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建 uts ns ;  ipc ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC,
	}

	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// 测试 pid 命名空间
func testPid() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建 uts ns ;  ipc ns ;pic ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID,
	}

	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
func testMount() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建 uts ns ;  ipc ns ;pic ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
func testUser() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建 uts ns ;  ipc ns ;pic ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1,
				Size:        1,
			},
		},
	}
	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
func testNet() {
	//exec.Command() 用于指定被fork出来的新进程内的初始命令，这里使用sh来执行
	//创建一个shell
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建 uts ns ;  ipc ns ;pic ns
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1,
				Size:        1,
			},
		},
	}
	// 将当前进程的 输入输出流 附加到 启动的进程上面
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
