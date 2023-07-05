package run

import (
	"cgroups"
	"containers"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Run(tty bool, cmdArray []string, res *cgroups.ResourceConfig, volume string) {
	parent, writePipe, containerBaseUrl := containers.NewParentProcess(tty, volume)
	if parent == nil {
		log.Println("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		fmt.Errorf("启动父进程失败:%v\n", err)
	}
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
	setUpMount()
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

// 初始化挂载点
func setUpMount() {
	//获取工作目录
	pwd, err := os.Getwd()
	if err != nil {
		_ = fmt.Errorf("获取当前工作目录失败：%v \n", err)
	}
	//挂载root目录
	err = pivotRoot(pwd)
	if err != nil {
		_ = fmt.Errorf("挂载root目录失败: %v \n ", err)
		return
	}

	//设置默认挂载参数 MS_NOEXEC本文将系统不允许运行其他程序 MS_NOEXEC 运行程序的时候
	// 不允许 set-user-id,set-group-id
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载 proc目录，使 容器有独立的proc目录
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		_ = fmt.Errorf("挂载 /proc 目录 失败: %v \n", err)
		return
	}
	//挂载 /dev 目录
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		_ = fmt.Errorf("挂载 /dev 目录 失败: %v \n", err)
		return
	}

}

func pivotRoot(containerRoot string) error {
	// 使用 mount --bind foo foo 方式 将 containerRoot重新挂载了一次
	//使得 cotainerRoot的文件系统和 宿主机的文件系统不同
	//这是 pivot_root的必须要求
	if err := syscall.Mount(containerRoot, containerRoot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}
	// 创建 cotainerRoot/.pivot_root 存储 old_root
	pivotDir := filepath.Join(containerRoot, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在cotainerRoot/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	// 根目录已经被替换
	if err := syscall.PivotRoot(containerRoot, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// 卸载掉 containerRoot/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹，如果不删除的话，就可以在容器里面访问到宿主机的根目录
	return os.Remove(pivotDir)
}
