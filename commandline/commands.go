package commandline

import (
	"cgroups"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"run"
)

func StartCommands() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "mydocker is a simple container runtime implementation"
	app.Commands = []cli.Command{RunCommand, InitCommand}
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
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
	},
	// 具体的执行命令
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing containers command")
		}
		// 获取要执行run的命令
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		// 获取tty参数
		tty := context.Bool("ti")
		res := &cgroups.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		// 获取卷挂载参数
		volume := context.String("v")
		run.Run(tty, cmdArray, res, volume)
		return nil
	},
}

// InitCommand 定义 init 命令，这是内部命令
var InitCommand = cli.Command{
	Name: "init",
	Usage: `Init containers process run user's process in containers. Do not call it outside 
		`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},
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
