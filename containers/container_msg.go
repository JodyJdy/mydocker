package containers

type ContainerInfo struct {
	Pid        string `json:"pid"`         //容器的init进程在宿主机上的进程id
	Id         string `json:"id"`          //容器id
	Name       string `json:"name"`        //容器name
	Command    string `json:"command"`     //容器内init进程的运行命令
	CreateTime string `json:"create_time"` //创建时间
	Status     string `json:"status"`      //容器的状态
	Volume     string `json:"volume"`      // 容器的卷挂载
	Image      string `json:"image"`       //使用镜像
}

// 定义目录相关的常量，存放信息

var (
	Running = "running"
	Stop    = "stoped"
	Exit    = "exited"
	// DefaultInfoLocation %s 是容器的标识
	DefaultInfoLocation  = "/var/run/mydocker/containers/%s/"
	AllContainerLocation = "/var/run/mydocker/containers/"
	ContainerConfigName  = "config.json"
	ContainerLogName     = "container.log"
)
