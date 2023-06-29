package commandline

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func Test() {
	app := cli.NewApp()
	app.Commands = []cli.Command{runCommand}
	app.Run(os.Args)
}

// docker run 命令
var runCommand = cli.Command{
	Name: "run",
	Usage: ` 创建一个容器，带有命名空间和cgroups的限制 mydocker run -ti  [command]
		`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},
	// 具体的执行命令
	Action: func(context *cli.Context) error {
		fmt.Println("hello world")
		return nil
	},
}
