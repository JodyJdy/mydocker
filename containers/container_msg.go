package containers

type ContainerInfo struct {
	Pid         string       `json:"pid"`         //容器的init进程在宿主机上的进程id
	Id          string       `json:"id"`          //容器id
	Name        string       `json:"name"`        //容器name
	Command     string       `json:"command"`     //容器内init进程的运行命令
	CreateTime  string       `json:"create_time"` //创建时间
	Status      string       `json:"status"`      //容器的状态
	Volume      []VolumeInfo `json:"volume"`      // 容器的卷挂载
	Image       string       `json:"image"`       //使用镜像
	BaseUrl     string       `json:"baseUrl"`     // 容器的文件系统目录
	SetCgroup   bool         `json:"setCgroup"`   //有无创建cgroup
	PortMapping []string     `json:"portMapping"` // 端口映射
	net         string       `json:"net"`         // 端口
}

type VolumeInfo struct {
	HostVolumePath      string `json:"hostVolumePath"`      //卷在宿主机上的路径
	ContainerPath       string `json:"containerPath"`       //容器中的相对路径
	ContainerPathInHost string `json:"containerPathInHost"` //容器中的路径在宿主机上的位置
	Anonymous           bool   `json:"anonymous"`           //是否是匿名卷
}

// 定义目录相关的常量，存放信息

var (
	Running = "running"
	Stop    = "stoped"
	Exit    = "exited"
	// ContainerInfoLocation %s 是容器的标识
	ContainerInfoLocation = "/var/run/mydocker/containers/%s/"
	AllContainerLocation  = "/var/run/mydocker/containers/"
	AllVolumeLocation     = "/var/run/mydocker/volumes/"
	VolumeInfoLocation    = "/var/run/mydocker/volumes/%s/"
	ContainerConfigName   = "config.json"
	ContainerLogName      = "container.log"
)
