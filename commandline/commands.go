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
	app.Usage = "mydocker is a simple container runtime implementation"
	app.Commands = []cli.Command{RunCommand, InitCommand, CommitCommand, PsCommand, LogCommand,
		ExecCommand, StopCommand, RemoveCommand, BuildBaseImageCommand, ImagesCommand, BuildImageCommand, NetworkCommand}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

// RunCommand docker run 命令
var RunCommand = cli.Command{
	Name: "run",
	Usage: ` Create a containers with namespace and cgroups limit mydocker run -ti  [command]
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
	Usage: `Init containers process run user's process in containers. Do not call it outside 
		`,
	// 具体的执行命令
	Action: func(context *cli.Context) error {
		log.Println("启动 init 进程")
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
	Usage: "commit a container into image",
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
					Usage: "subnet cidr",
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
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				networks.Init()
				networks.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container network",
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
