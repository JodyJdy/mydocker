package commandline

import (
	"cgroups"
	"containers"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"networks"
	"nsenter"
	"os"
	"run"
)

func StartCommands() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "容器运行时实现"
	app.Commands = []cli.Command{RunCommand, InitCommand, CommitCommand, PsCommand, LogCommand,
		ExecCommand, StopCommand, RemoveCommand, BuildBaseImageCommand, ImagesCommand, BuildImageCommand, NetworkCommand, PortMappingCommand, SaveCommand}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

// RunCommand docker run 命令
var RunCommand = cli.Command{
	Name: "run",
	Usage: ` 启动容器
		`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringSliceFlag{
			Name:  "v",
			Usage: "volume，可挂载多个",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "后台运行进程",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "设置环境变量",
		},
		cli.StringFlag{
			Name:  "image",
			Usage: "镜像id前缀 或者 镜像名称",
		},

		cli.StringFlag{
			Name:  "net",
			Usage: "指定容器所属的网络 -net host 和宿主机共享网络 -net container:容器标识 和容器共享网络 -net bridgeName 指定网络 ",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port 映射",
		},
	},
	// 具体的执行命令
	Action: func(context *cli.Context) error {
		// 获取要执行run的命令
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		// 获取tty参数
		tty := context.Bool("ti")
		// 获取 detach 参数
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("ti 和 d 不能同时使用")
		}
		res := &cgroups.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		// 获取卷挂载参数
		volumes := context.StringSlice("v")
		// 获取容器名称
		containerName := context.String("name")
		//获取容器名称
		envSlice := context.StringSlice("e")
		// 获取使用的镜像
		imageId := context.String("image")
		// 端口映射
		portmapping := context.StringSlice("p")
		// 所在的网络
		network := context.String("net")
		if imageId == "" {
			fmt.Println("镜像id不能为空")
			return nil
		}
		run.Run(tty, cmdArray, res, volumes, containerName, envSlice, imageId, portmapping, network)
		return nil
	},
}

// InitCommand 定义 init 命令，这是内部命令
var InitCommand = cli.Command{
	Name: "init",
	Usage: `内部用于初始化容器进程，不能从外部访问
		`,
	// 具体的执行命令
	Action: func(context *cli.Context) error {
		log.Println("容器 init 中 ")
		err := run.RunContainerInitProcess()
		if err != nil {
			return err
		}
		return nil
	},
}

// CommitCommand 镜像提交命令 @Todo 打包镜像先不做
var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "提交容器为镜像",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("缺少容器名称和镜像名称")
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		fmt.Println(containerName, imageName)
		return nil
	},
}

var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "列出所有容器",
	Action: func(context *cli.Context) error {
		run.Ps()
		return nil
	},
}

var LogCommand = cli.Command{
	Name:  "logs",
	Usage: "打印容器日志",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("缺少容器名称或者标识")
		}
		run.Log(context.Args()[0])
		return nil
	},
}

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "在容器中执行命令",
	Action: func(context *cli.Context) error {
		// 说明当前是fork的进程，环境变量已经设置好， c语言的代码也已经执行了
		if os.Getenv(containers.ENV_EXEC_PID) != "" {
			fmt.Printf("pid callback pid %d\n", os.Getpid())
			nsenter.SetNs()
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("缺少容器id标识或者执行命令\n")
		}
		containerId := context.Args()[0]
		//获取执行的命令
		var commandArray []string
		// tail函数或获取除第一个参数以外的参数
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		//执行命令
		run.Exec(containerId, commandArray)
		return nil
	},
}
var StopCommand = cli.Command{
	Name:  "stop",
	Usage: "停止容器",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("缺少容器名称或标识")
		}
		run.Stop(context.Args()[0])
		return nil
	},
}
var RemoveCommand = cli.Command{
	Name:  "remove",
	Usage: "删除容器",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("缺少容器名称或标识")
		}
		run.Remove(context.Args()[0])
		return nil
	},
}

var BuildBaseImageCommand = cli.Command{
	Name:  "buildBase",
	Usage: "构建基础镜像",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("缺少基础镜像tar包路径")
		}
		containers.BuildBaseImage(context.Args()[0])
		return nil
	},
}
var ImagesCommand = cli.Command{
	Name:  "images",
	Usage: "展示镜像",
	Action: func(context *cli.Context) error {
		containers.ListImageInfo()
		return nil
	},
}
var BuildImageCommand = cli.Command{
	Name:  "build",
	Usage: "构建镜像",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t",
			Usage: "镜像标签",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "docker file路径",
		},
	},
	Action: func(context *cli.Context) {
		containers.BuildImage(context.String("t"), context.String("f"))
	},
}
var NetworkCommand = cli.Command{
	Name:  "network",
	Usage: "创建容器网络",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "创建网络",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "创建网络驱动",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "子网cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("缺少网络名称")
				}
				networks.Init()
				err := networks.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("create network error: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "列出创建的网络",
			Action: func(context *cli.Context) error {
				networks.Init()
				networks.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "删除网络",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("Missing network name")
				}
				networks.Init()
				err := networks.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("remove network error: %+v", err)
				}
				return nil
			},
		},
	},
}

var PortMappingCommand = cli.Command{
	Name:  "portmap",
	Usage: "管理端口映射",
	Subcommands: []cli.Command{
		{
			Name:  "start",
			Usage: "启动端口映射服务",
			Action: func(context *cli.Context) error {
				networks.StartPortMapping()
				return nil
			},
		},
		{
			Name:  "forward",
			Usage: "端口转发",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "p",
					Usage: "手动配置端口映射",
				},
				cli.StringSliceFlag{
					Name:  "d",
					Usage: "删除端口映射",
				},
			},
			Action: func(context *cli.Context) error {
				// 端口映射
				portMappings := context.StringSlice("p")
				// 删除端口映射
				removePortMappings := context.StringSlice("d")
				networks.SendPortMapping(portMappings, removePortMappings)
				return nil
			},
		},
	},
}

var SaveCommand = cli.Command{
	Name:  "save",
	Usage: "保存容器为tar文件",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "o",
			Usage: "保存的文件名,以tar结尾",
		},
		cli.StringFlag{
			Name:  "c",
			Usage: "容器标识",
		},
	},
	Action: func(context *cli.Context) {
		containers.SaveContainer(context.String("c"), context.String("o"))
	},
}
