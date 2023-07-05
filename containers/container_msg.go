package containers

type ContainerInfo struct {
	Pid        string `json:"pid"`         //容器的init进程在宿主机上的进程id
	Id         string `json:"id"`          //容器id
	Name       string `json:"name"`        //容器name
	Command    string `json:"command"`     //容器内init进程的运行命令
	CreateTime string `json:"create_time"` //创建时间
	Status     string `json:"status"`      //容器的状态
}

// 定义目录相关的常量，存放信息

var (
	Running string = "running"
	Stop    string = "stoped"
	Exit    string = "exited"
	// DefaultInfoLocation %s 是容器的标识
	DefaultInfoLocation  string = "/var/run/mydocker/containers/%s/"
	AllContainerLocation string = "/var/run/mydocker/containers/"
	ConfigName           string = "config.json"
)
