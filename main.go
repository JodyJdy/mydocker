package main

import (
	"commandline"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	commandline.StartCommands()
}

// worked
func test1() {
	cmd := exec.Command("sh", "-c", "sleep 9")
	err := cmd.Run()
	fmt.Println(cmd.Process.Pid)
	fmt.Println(err)
}
func test2() {
	// 后面的sleep 900 是参数， 不能有 双引号
	cmdArray := []string{"sh", "-c", "sleep 900 "}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Printf("Exec loop path error %v\n", err)
	}
	fmt.Printf("Find path %s\n", path)
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
}

func test3() {
	//commandline.StartCommands()
	cmdArray := []string{"sh", "-c", "./x.sh"}
	//cmdArray := []string{"./x.sh"}

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		fmt.Printf("Exec loop path error %v\n", err)
	}
	fmt.Printf("Find path %s\n", path)
	// 当前处于父进程中， exec 会执行cmd，将cmd对应的进程代替父进程
	//也就是说容器中 pid =1的进程会是 cmd对应的进程
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		fmt.Println(err.Error())
	}
}
func test4() {
	// 两种方式都可以
	cmd := exec.Command("sh", "-c", "/mnt/c/Users/Admin/GolandProjects/mydocker/x.sh")
	//cmd := exec.Command("sh", "-c", "./x.sh")
	err := cmd.Run()
	fmt.Println(cmd.Process.Pid)
	fmt.Println(err)
}
