package containers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type CommandArray struct {
	// 命令的执行类型
	ShellType bool `json:"shellType"`
	// 命令数组
	Cmds []string `json:"cmds"`
}

func SaveCommand(array *CommandArray, file *os.File) {
	contents, err := json.Marshal(array)
	if err != nil {
		fmt.Println("序列化command array 失败")
		return
	}
	_, err = file.Write(contents)
	// 管道要关闭，才能让另一个进程知道已经结束了
	defer file.Close()
	if err != nil {
		fmt.Println("命令写入管道失败")
		return
	}
}
func LoadCommand(array *CommandArray, file *os.File) *CommandArray {
	contents, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("读取命令行内容失败")
	}
	err = json.Unmarshal(contents, &array)
	if err != nil {
		fmt.Println("命令行内容序列化为对象失败")
	}
	return array
}

func ResolveCmd(cmdArray []string, imageId string, tty bool) *CommandArray {
	info, err := GetImageInfo(imageId)
	if err != nil {
		fmt.Errorf("获取镜像失败: %s, 原因: %v", cmdArray, err)
	}
	result := CommandArray{}
	// tty就不会执行后台进程
	if tty {
		result.Cmds = cmdArray
		result.ShellType = false
		return &result
	}

	if info.EntryPointShellType {
		// 不可覆盖
		result.Cmds = []string{"sh", "-c", strings.Join(info.EntryPoint, " ")}
		result.ShellType = true
	} else {
		// 可被覆盖
		result.ShellType = false
		result.Cmds = info.EntryPoint
		// 未指定 EntryPoint
		if len(result.Cmds) == 0 {
			// 用户未输入，则使用原有的cmd
			if len(cmdArray) == 0 {
				result.ShellType = info.CMDShellType
				if info.CMDShellType {
					result.Cmds = []string{"sh", "-c", strings.Join(info.CMD, " ")}
				} else {
					result.Cmds = info.CMD
				}
			} else {
				// 被覆盖
				result.ShellType = true
				result.Cmds = []string{"sh", "-c", strings.Join(cmdArray, " ")}
			}
		} else {
			// cmd 和 用户输入的指令都是作为参数处理
			if (len(cmdArray)) != 0 {
				result.Cmds = append(result.Cmds, strings.Join(cmdArray, " "))
			} else {
				result.Cmds = append(result.Cmds, strings.Join(info.CMD, " "))
			}
		}

	}
	return &result

}
