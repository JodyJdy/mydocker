package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"sysv_mq"
	"time"
)

func main() {
	//commandline.StartCommands()
	mq, err := sysv_mq.NewMessageQueue(&sysv_mq.QueueConfig{
		Key:     0xDEADBEEF,
		MaxSize: 1024,
		Mode:    sysv_mq.IPC_CREAT | 0600,
	})
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		for {
			err = mq.SendBytes([]byte("Hello World"), 1, sysv_mq.IPC_NOWAIT)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}()

	for {
		response, _, err := mq.ReceiveBytes(0, sysv_mq.IPC_NOWAIT)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(response))
		}
		time.Sleep(time.Millisecond * 1000)
	}
	ch := make(chan interface{}, 1)
	<-ch

}

// worked
func test1() {
	cmd := exec.Command("sh", "-c", "touch x.txt")
	err := cmd.Start()
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
