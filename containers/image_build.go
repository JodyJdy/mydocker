package containers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"
	"time"
)

func BuildBaseImage(imageTarUrl string) {
	if !FileExist(imageTarUrl) {
		fmt.Printf("文件不存在:%s\n", imageTarUrl)
	}
	// 删除原先的基础镜像信息
	err := os.RemoveAll(BaseImageUrl)
	if err != nil {
		fmt.Printf("删除基础镜像信息失败: %v\n", err)
		return
	}
	err = os.MkdirAll(BaseImageLayerLocation, 0622)
	if err != nil {
		fmt.Printf("创建目录失败 %s  %v\n", BaseImageLayerLocation, err)
		return
	}

	storeBaseImageInfo()
	// 解压文件，tar包
	if _, err := exec.Command("tar", "--strip-components", "1", "-xvf", imageTarUrl, "-C", BaseImageLayerLocation).CombinedOutput(); err != nil {
		fmt.Printf("Unbar dir %s error %v\n", BaseImageLayerLocation, err)
		return
	}

}

// 存储基础镜像信息
func storeBaseImageInfo() {
	info := ImageInfo{
		Name:                GetBaseImageId(),
		Id:                  GetBaseImageId(),
		CreateTime:          time.Now().Format("2006-01-02 15:04:05"),
		EntryPoint:          []string{"sh", "-c"},
		EntryPointShellType: false,
		CMD:                 []string{"echo I am base image"},
		CMDShellType:        false,
		Version:             "",
		Volume:              []string{"/a", "/b", "/home/jdy"},
	}
	recordImageInfo(&info)
}

func recordImageInfo(info *ImageInfo) {
	// 序列化为字符串
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		fmt.Printf("记录镜像信息失败: %v", err)
		return
	}
	jsonStr := string(jsonBytes)
	// 镜像信息记录的路径
	dirUrl := fmt.Sprintf(ImageInfoLocation, info.Id)
	// 尝试创建路径
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		fmt.Printf("创建路径%s 失败: %v", dirUrl, err)
	}
	fileName := dirUrl + ConfigName
	//删除旧的文件，如果存在的话
	_ = os.Remove(fileName)
	// 创建文件
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		fmt.Printf("创建文件失败%s 失败: %v", fileName, err)
		return
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		fmt.Printf("写入镜像信息失败: %v", err)
	}
}
func ListImageInfo() {
	// 返回所有容器的目录
	imageDirs, err := os.ReadDir(AllImageLocation)
	if err != nil {
		fmt.Errorf("read dir %s error %v", AllImageLocation, err)
		return
	}
	// 记录所有容器的对象
	var imageInfos []*ImageInfo
	for _, containerDir := range imageDirs {
		tmpContainer, err := ReadImageInfo(containerDir)
		if err != nil {
			fmt.Errorf("Get container info error %v", err)
			continue
		}
		imageInfos = append(imageInfos, tmpContainer)
	}
	// 格式化并输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tVERSION\tFROM\tEXPOSE\tCREATED\n")
	for _, item := range imageInfos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Version,
			item.From,
			item.Expose,
			item.CreateTime)
	}
	if err := w.Flush(); err != nil {
		fmt.Errorf("Flush error %v\n", err)
		return
	}
}

func ReadImageInfo(imageDir os.DirEntry) (*ImageInfo, error) {
	dir := fmt.Sprintf(ImageInfoLocation, imageDir.Name())
	imageInfoDir := dir + ConfigName
	content, err := os.ReadFile(imageInfoDir)
	if err != nil {
		fmt.Errorf("read image Dir %s error %v", imageInfoDir, err)
		return nil, err
	}
	var info ImageInfo
	if err := json.Unmarshal(content, &info); err != nil {
		fmt.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &info, nil
}
func GetImageInfo(imageId string) (*ImageInfo, error) {
	dir := fmt.Sprintf(ImageInfoLocation, imageId)
	imageInfoFile := dir + ConfigName
	content, err := os.ReadFile(imageInfoFile)
	if err != nil {
		fmt.Errorf("read image info %s error %v \n", imageInfoFile, err)
		return nil, err
	}
	var info ImageInfo
	if err := json.Unmarshal(content, &info); err != nil {
		fmt.Errorf("json unmarshal error %v", err)
		return nil, err
	}
	return &info, nil
}

func BuildImage(tag string, dockerFile string) {
	if !FileExist(dockerFile) {
		fmt.Printf("docker file 不存在: %s", dockerFile)
		return
	}
	//获取镜像id
	imageId := ImageId()
	//创建镜像目录
	if err := os.MkdirAll(ImageLayerLocation+imageId, 0622); err != nil {
		fmt.Printf("创建镜像目录失败: %v", err)
		return
	}
	info := &ImageInfo{
		Id:         imageId,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	fmt.Println(info)
	d := &DockerFile{
		// 默认的工作目录
		WorkDir: "/",
	}
	file, _ := os.Open(dockerFile)
	var containerId = new(string)
	r := bufio.NewReader(file)
	for {
		s, _, err := r.ReadLine()
		// 读取结束
		if err != nil {
			break
		}
		line := string(s)
		fmt.Println(line)
		switch {
		case strings.HasPrefix(line, FROM):
			d.from(line, containerId)
			break
		case strings.HasPrefix(line, RUN):
			d.run(line, *containerId)
			break
		case strings.HasPrefix(line, ADD):
			d.add(line)
			break
		case strings.HasPrefix(line, COPY):
			d.copy(line)
			break
		case strings.HasPrefix(line, EXPOSE):
			d.expose(line)
			break
		case strings.HasPrefix(line, ENV):
			d.env(line)
			break
		case strings.HasPrefix(line, CMD):
			d.cmd(line)
			break
		case strings.HasPrefix(line, ENTRYPOINT):
			d.entrypoint(line)
			break
		case strings.HasPrefix(line, VOLUME):
			d.volume(line)
			break
		case strings.HasPrefix(line, WORKDIR):
			d.workDir(line)
			break
		default:
			continue
		}
	}
	recordImageInfo(info)
}
func (d *DockerFile) from(f string, containerId *string) {
	f = strings.TrimPrefix(f, FROM)
	d.From = strings.Trim(f, " ")
	*containerId = BuildFrom(d.From)
}
func (d *DockerFile) run(r string, containerId string) {
	r = strings.TrimPrefix(r, RUN)
	r, b := isArrayType(r)
	if b {
		BuildRun(containerId, parseArray(r))
	} else {
		BuildRun(containerId, parseCommandLine(r))
	}
}
func (d *DockerFile) add(a string) {
	a = strings.TrimPrefix(a, ADD)
	a, b := isArrayType(a)
	var list []string
	if b {
		list = parseArray(a)
	} else {
		list = parseCommandLine(a)
	}
	fmt.Println(list)
}
func (d *DockerFile) copy(c string) {
	c = strings.TrimPrefix(c, COPY)
	c, b := isArrayType(c)
	var list []string
	if b {
		list = parseArray(c)
	} else {
		list = parseCommandLine(c)
	}
	fmt.Println(list)
}
func (d *DockerFile) expose(e string) {
	e = strings.TrimPrefix(e, EXPOSE)
	// 端口列表
	ports := parseCommandLine(e)
	d.Expose = ports
}
func (d *DockerFile) env(e string) {
	e = strings.TrimPrefix(e, ENV)
	e, b := isArrayType(e)
	var list []string
	if b {
		list = parseArray(e)
	} else {
		list = parseCommandLine(e)
	}
	fmt.Println(list)
}
func (d *DockerFile) cmd(c string) {
	c = strings.TrimPrefix(c, CMD)
	c, b := isArrayType(c)
	if b {
		d.CMD = parseArray(c)
		d.EntryPointShellType = false
	} else {
		d.CMD = parseCommandLine(c)
		d.EntryPointShellType = true
	}
}
func (d *DockerFile) entrypoint(e string) {
	e = strings.TrimPrefix(e, ENTRYPOINT)
	e, b := isArrayType(e)
	if b {
		d.EntryPoint = parseArray(e)
		d.EntryPointShellType = false
	} else {
		d.EntryPoint = parseCommandLine(e)
		d.EntryPointShellType = true
	}
}
func (d *DockerFile) volume(v string) {
	v = strings.TrimPrefix(v, VOLUME)
	d.Volumes = parseCommandLine(v)
}
func (d *DockerFile) workDir(w string) {
	w = strings.TrimPrefix(w, WORKDIR)
	d.WorkDir = path.Clean(w)
}

// 判断是否是数组类型
func isArrayType(s string) (string, bool) {
	s = strings.Trim(s, " ")
	if s[0] == '[' {
		return s, true
	}
	return s, false
}

// 解析 ["a","b"] 类型
func parseArray(s string) []string {
	le := len(s)
	var array []string
	for i := 0; i < le; {
		if s[i] == '[' || s[i] == ',' {
			i++
			continue
		}
		if s[i] == '"' && i+1 < le {
			if s[i+1] == '"' {
				i += 2
				continue
			}
			j := i + 1
			for ; j < le; j++ {
				if s[j] == '"' && s[j-1] != '\\' {
					array = append(array, s[i+1:j])
					break
				}
			}
			i = j + 1
			continue
		}
		if s[i] == ']' {
			break
		}
		i++

	}
	fmt.Printf("解析内容是： %v\n", array)
	return array
}

// 解析  非数组类型
func parseCommandLine(s string) []string {
	le := len(s)
	var cmd []string
	for i := 0; i < le; {
		if s[i] == ' ' {
			for i < le {
				if s[i] == ' ' {
					i++
				} else {
					break
				}
			}
			continue
		}
		if s[i] == '"' && i+1 < le {
			if s[i+1] == '"' {
				i += 2
				continue
			}
			j := i + 1
			for ; j < le; j++ {
				if s[j] == '"' && s[j-1] != '\\' {
					cmd = append(cmd, s[i+1:j])
					break
				}
			}
			i = j + 1
			continue
		}
		j := i + 1
		for j < le {
			if s[j] == ' ' || s[j] == '"' {
				break
			} else {
				j++
			}
		}
		cmd = append(cmd, s[i:j])
		i = j
	}
	fmt.Printf("解析内容是： %v\n", cmd)
	return cmd
}
