package containers

type ImageInfo struct {
	Id                  string   `json:"id"`                  //镜像id
	Name                string   `json:"name"`                //镜像name
	Version             string   `json:"version"`             // 版本号
	CreateTime          string   `json:"create_time"`         //创建时间
	Env                 []string `json:"env"`                 //环境变量
	Volume              []string `json:"volume"`              // 匿名卷挂载
	Expose              []string `json:"expose"`              // 暴露端口
	Label               []string `json:"label"`               //标签信息
	From                string   `json:"from"`                //基础镜像
	EntryPoint          []string `json:"entryPoint"`          // entryPoint
	EntryPointShellType bool     `json:"entryPointShellType"` //entrypoint是shell类型还是exec类型
	CMD                 []string `json:"cmd"`                 // CMD
	CMDShellType        bool     `json:"CMDShellType"`        // cmd是shell类型还是exec类型
	WorkDir             string   `json:"workDir"`             // workDir
}

var (
	// ImageInfoLocation %s 是镜像的标识
	ImageInfoLocation = "/var/run/mydocker/images/%s/"
	// ImageLayerLocation 镜像目录
	ImageLayerLocation     = AllImageLocation + "%s/layer/"
	AllImageLocation       = "/var/run/mydocker/images/"
	BaseImageUrl           = AllImageLocation + "base/"
	BaseImageLayerLocation = BaseImageUrl + "layer/"
	// ImageConfigName 存储镜像信息
	ConfigName = "config.json"
)

// DockerFile 解析DockerFile,解析时，有些是直接执行的，有些是需要留档的
type DockerFile struct {
	// 暂时使用镜像id
	From                string
	Expose              []string
	Env                 []string
	EntryPoint          []string // entryPoint
	EntryPointShellType bool     //entrypoint是shell类型还是exec类型
	CMD                 []string // CMD
	CMDShellType        bool     // cmd是shell类型还是exec类型
	Volumes             []string
	// 工作目录
	WorkDir string
}

const FROM = "FROM"
const RUN = "RUN"
const ADD = "ADD"
const COPY = "COPY"
const EXPOSE = "EXPOSE"
const ENV = "ENV"
const CMD = "CMD"
const ENTRYPOINT = "ENTRYPOINT"
const VOLUME = "VOLUME"
const WORKDIR = "WORKDIR"
