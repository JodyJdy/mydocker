package run

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	var cmdArray CommandArray
	err = json.Unmarshal(contents, &cmdArray)
	if err != nil {
		fmt.Println("命令行内容序列化为对象失败")
	}
	return &cmdArray
}
