package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

func main() {
	testCgroup()
}

// 搭载了 memory subsystem的hierarchy的根目录位置
var cgroupMemoryHierachyMount = "/sys/fs/cgroup/memory"

func testCgroup() {

	// 表示是 fork出来的进程，执行了 /proc/self/exe创建出来的
	// 为容器中的进程
	if os.Args[0] == "/proc/self/exe" {
		//获取容器进程id
		fmt.Println("current pid :", syscall.Getpid())
		//执行占内存的操作，用于判断，内存有没有被限制
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// 执行cmd.Run  用于真正的执行命令
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// fork 当前进程
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// start也可以执行命令，和Run的区别是，Run会等待命令执行完毕，start不会
	//因为接下来要创建cgroup所以不能使用Run
	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	} else {
		//打印fork出的进程的id
		fmt.Println(cmd.Process.Pid)
		//创建cgroup
		err := os.Mkdir(path.Join(cgroupMemoryHierachyMount, "testmemorylimit"), 0755)
		if err != nil {
			return
		}
		//将容器进程写入到cgroup中
		err = os.WriteFile(path.Join(cgroupMemoryHierachyMount, "testmemorylimit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		if err != nil {
			return
		}
		//限制cgroup进程占用
		err = os.WriteFile(path.Join(cgroupMemoryHierachyMount, "testmemorylimit", "memory.limit_in_bytes"), []byte("100m"), 0644)
		if err != nil {
			return
		}
	}
	//等待命令执行完毕
	err := cmd.Wait()
	if err != nil {
		return
	}
}
