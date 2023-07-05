package containers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

// NewParentProcess 创建一个父进程， 父进程的目的是
// 真正的执行cmd，并用cmd 对应的进程替换自身
func NewParentProcess(containerId string, tty bool, volume string) (*exec.Cmd, *os.File, string) {
	// 容器工作目录 /var/run/mydocker/containers/容器id/
	containerBaseUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		_ = fmt.Errorf("new pipe error %v", err)
		return nil, nil, ""
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
	} else {
		// 生产容器对应目录的container.log文件
		if err := os.MkdirAll(containerBaseUrl, 0622); err != nil {
			fmt.Printf("创建目录 %s 失败 %v", containerBaseUrl, err)
			return nil, nil, ""
		}
		logFilePath := containerBaseUrl + ContainerLogName
		logFile, err := os.Create(logFilePath)
		if err != nil {
			fmt.Printf("创建日志文件 %s 失败 %v", logFilePath, err)
			return nil, nil, ""
		}
		// 将进程的输出重定向到logFile中，访问这个文件，就能读取到日志
		cmd.Stdout = logFile
	}
	// 用于读取 管道中的命令
	cmd.ExtraFiles = []*os.File{readPipe}
	// 工作目录，为 overlay文件系统中的 merge目录 ,容器进程，会以merged目录作为根目录运行
	cmd.Dir = NewWorkSpace(containerBaseUrl, volume)
	return cmd, writePipe, containerBaseUrl
}

// NewPipe 创建管道对象
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// RecordContainerInfo 记录容器信息
func RecordContainerInfo(id string, containerPID int, cmdArray []string, containerName string, volume string) {
	//获取容器创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	// 调整容器名称，用户不传时的默认值
	if containerName == "" {
		containerName = id
	}
	containerInfo := &ContainerInfo{
		Id:         id,
		CreateTime: createTime,
		Pid:        strconv.Itoa(containerPID),
		Command:    strings.Join(cmdArray, ""),
		Status:     Running,
		Name:       containerName,
		Volume:     volume,
	}
	recordContainerInfo(containerInfo)
}
func recordContainerInfo(info *ContainerInfo) {
	// 序列化为字符串
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		fmt.Printf("记录容器信息失败: %v", err)
		return
	}
	jsonStr := string(jsonBytes)
	// 容器信息记录的路径
	dirUrl := fmt.Sprintf(DefaultInfoLocation, info.Id)
	// 尝试创建路径
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		fmt.Printf("创建路径%s 失败: %v", dirUrl, err)
	}
	fileName := dirUrl + ConfigName
	//删除旧的文件，如果存在的话
	_ = os.Remove(fileName)
	// 创建文件
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		fmt.Printf("创建文件失败%s 失败: %v", fileName, err)
		return
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		fmt.Printf("写入容器信息失败: %v", err)
	}
}

// DeleteContainerInfo 删除容器信息
func DeleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		fmt.Printf("删除目录：%s失败 %v", dirUrl, err)
	}
}

func ListContainerInfo() {
	// 返回所有容器的目录
	containerDirs, err := os.ReadDir(AllContainerLocation)
	if err != nil {
		fmt.Errorf("read dir %s error %v", AllContainerLocation, err)
		return
	}
	// 记录所有容器的对象
	var containers []*ContainerInfo
	for _, containerDir := range containerDirs {
		tmpContainer, err := ReadContainerInfo(containerDir)
		if err != nil {
			fmt.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 格式化并输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}
	if err := w.Flush(); err != nil {
		fmt.Errorf("Flush error %v", err)
		return
	}
}
func ReadContainerInfo(containerDir os.DirEntry) (*ContainerInfo, error) {
	dir := fmt.Sprintf(DefaultInfoLocation, containerDir.Name())
	containerInfoFile := dir + ConfigName
	content, err := os.ReadFile(containerInfoFile)
	if err != nil {
		fmt.Errorf("read containerDir %s error %v", containerInfoFile, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		fmt.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
func GetContainerInfo(containerId string) (*ContainerInfo, error) {
	dir := fmt.Sprintf(DefaultInfoLocation, containerId)
	containerInfoFile := dir + ConfigName
	content, err := os.ReadFile(containerInfoFile)
	if err != nil {
		fmt.Errorf("read containerDir %s error %v", containerInfoFile, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		fmt.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
func GetContainerPid(containerId string) string {
	info, _ := GetContainerInfo(containerId)
	return info.Pid
}

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func ExecContainer(containerId string, cmdArray []string) {
	pid := GetContainerPid(containerId)
	//拼接命令行
	cmdStr := strings.Join(cmdArray, " ")
	fmt.Printf("容器进程pid是%s,执行命令%s \n", pid, cmdStr)
	//再次调用自身
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//设置环境变量， 用于 c 相关的代码判断是否执行，以及作为c执行的参数
	err := os.Setenv(ENV_EXEC_PID, pid)
	if err != nil {
		fmt.Println("设置环境变量失败")
		return
	}
	err = os.Setenv(ENV_EXEC_CMD, cmdStr)
	if err != nil {
		fmt.Println("设置环境变量失败")
		return
	}
	if err := cmd.Run(); err != nil {
		fmt.Printf("exec contaienr %s error", containerId, err)
	}
}

func StopContainer(containerId string) {
	info, err := GetContainerInfo(containerId)
	if err != nil {
		fmt.Printf("获取容器:%s 进程pid,失败 %v\n", containerId, err)
	}
	pid, _ := strconv.Atoi(info.Pid)
	// 调用 kill
	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		fmt.Printf("关闭容器失败 %v\n", err)
		return
	}
	// 修改容器状态
	info.Pid = ""
	info.Status = Stop
	//记录到容器信息
	recordContainerInfo(info)
}
func RemoveContainer(containerId string) {
	//获取容器信息
	info, err := GetContainerInfo(containerId)
	if err != nil {
		fmt.Printf("获取容器:%s 进程,失败 %v\n", containerId, err)
		return
	}
	if info.Status != Stop {
		fmt.Println("只能删除停止的容器")
		return
	}
	baseUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	DeleteWorkSpace(baseUrl, info.Volume)
	DeleteContainerInfo(info.Id)
}
