package containers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"
	"time"
)

func BuildBaseImage(imageTarUrl string) {
	if !FileExist(imageTarUrl) {
		log.Printf("文件不存在:%s\n", imageTarUrl)
	}
	// 删除原先的基础镜像信息
	err := os.RemoveAll(BaseImageUrl)
	if err != nil {
		log.Printf("删除基础镜像信息失败: %v\n", err)
		return
	}
	err = os.MkdirAll(BaseImageLayerLocation, 0622)
	if err != nil {
		log.Printf("创建目录失败 %s  %v\n", BaseImageLayerLocation, err)
		return
	}
	storeBaseImageInfo()
	// 解压文件，tar包
	if _, err := exec.Command("tar", "-xvf", imageTarUrl, "-C", BaseImageLayerLocation).CombinedOutput(); err != nil {
		log.Printf("解压目录 dir %s 失败 %v\n", BaseImageLayerLocation, err)
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
		Volume:              []string{},
		WorkDir:             "/",
	}
	recordImageInfo(&info)
}

func recordImageInfo(info *ImageInfo) {
	// 序列化为字符串
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		log.Printf("记录镜像信息失败: %v", err)
		return
	}
	jsonStr := string(jsonBytes)
	// 镜像信息记录的路径
	dirUrl := fmt.Sprintf(ImageInfoLocation, info.Id)
	// 尝试创建路径
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Printf("创建路径%s 失败: %v", dirUrl, err)
	}
	fileName := dirUrl + ImageConfigName
	//删除旧的文件，如果存在的话
	_ = os.Remove(fileName)
	// 创建文件
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Printf("创建文件失败%s 失败: %v", fileName, err)
		return
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Printf("写入镜像信息失败: %v", err)
	}
}
func ListImageInfo() {
	// 格式化并输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tVERSION\tFROM\tEXPOSE\tCREATED\n")
	for _, item := range GetImageInfoList() {
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
func GetImageInfoList() []*ImageInfo {
	// 返回所有容器的目录
	imageDirs, err := os.ReadDir(AllImageLocation)
	if err != nil {
		log.Printf("读取目录失败 %s %v", AllImageLocation, err)
		return nil
	}
	// 记录所有容器的对象
	var imageInfos []*ImageInfo
	for _, containerDir := range imageDirs {
		tmpContainer, err := ReadImageInfo(containerDir)
		if err != nil {
			log.Printf("获取容器信息失败 %v", err)
			continue
		}
		imageInfos = append(imageInfos, tmpContainer)
	}
	return imageInfos

}

func ReadImageInfo(imageDir os.DirEntry) (*ImageInfo, error) {
	dir := fmt.Sprintf(ImageInfoLocation, imageDir.Name())
	imageInfoDir := dir + ImageConfigName
	content, err := os.ReadFile(imageInfoDir)
	if err != nil {
		log.Printf("读取镜像目录 Dir %s 失败 %v", imageInfoDir, err)
		return nil, err
	}
	var info ImageInfo
	if err := json.Unmarshal(content, &info); err != nil {
		log.Printf("json 反序列失败 %v\n", err)
		return nil, err
	}
	return &info, nil
}
func GetImageInfo(imageId string) (*ImageInfo, error) {
	dir := fmt.Sprintf(ImageInfoLocation, imageId)
	imageInfoFile := dir + ImageConfigName
	content, err := os.ReadFile(imageInfoFile)
	if err != nil {
		log.Printf("读取镜像信息失败 %s %v \n", imageInfoFile, err)
		return nil, err
	}
	var info ImageInfo
	if err := json.Unmarshal(content, &info); err != nil {
		log.Printf("json 反序列失败 %v\n", err)
		return nil, err
	}
	return &info, nil
}

func readDockerFile(dockerFile string) ([]string, error) {
	if !FileExist(dockerFile) {
		return nil, fmt.Errorf("docker file 不存在: %s", dockerFile)
	}

	file, err := os.Open(dockerFile)
	if err != nil {
		return nil, fmt.Errorf("打开 docker file 失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result []string
	var currentLine strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 续行处理（以 \ 结尾）
		if strings.HasSuffix(line, `\`) {
			currentLine.WriteString(strings.TrimSuffix(line, `\`))
			currentLine.WriteString(" ") // 保留一个空格，避免拼接时连在一起
			continue
		}

		// 最后一行或完整行
		currentLine.WriteString(line)
		result = append(result, strings.TrimSpace(currentLine.String()))
		currentLine.Reset()
	}

	// 处理扫描错误
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取 docker file 出错: %w", err)
	}

	// 如果还有残留未加入（可能文件结尾是续行）
	if currentLine.Len() > 0 {
		result = append(result, strings.TrimSpace(currentLine.String()))
	}

	return result, nil
}
func BuildImage(tag string, dockerFile string) {
	lines, err := readDockerFile(dockerFile)

	if err != nil {
		log.Fatalf("dockerfile解析失败 %v \n", err)
	}
	// 初始化 镜像信息
	info := initImageInfo(tag)
	// 初始化 dockerfile信息
	d := initDockerFile()
	for _, line := range lines {
		log.Println(line)
		switch {
		case strings.HasPrefix(line, FROM):
			d.from(line)
		case strings.HasPrefix(line, RUN):
			d.run(line)
		case strings.HasPrefix(line, ADD):
			d.add(line)
		case strings.HasPrefix(line, COPY):
			d.copy(line)
		case strings.HasPrefix(line, EXPOSE):
			d.expose(line)
		case strings.HasPrefix(line, ENV):
			d.env(line)
		case strings.HasPrefix(line, CMD):
			d.cmd(line)
		case strings.HasPrefix(line, ENTRYPOINT):
			d.entrypoint(line)
		case strings.HasPrefix(line, VOLUME):
			d.volume(line)
		case strings.HasPrefix(line, WORKDIR):
			d.workDir(line)
		default:
			continue
		}
	}
	//创建镜像目录
	if err := os.MkdirAll(fmt.Sprintf(ImageLayerLocation, info.Id), 0622); err != nil {
		log.Fatalf("创建镜像目录失败: %v", err)
	}
	//信息拷贝到 镜像信息中
	d.copy2ImageInfo(info)
	//记录镜像的信息
	recordImageInfo(info)
	// 拷贝镜像的Upper内容到layer，新的镜像就完成了
	Copy(path.Join(d.Info.BaseUrl, "upper")+"/*", fmt.Sprintf(ImageLayerLocation, info.Id))
	// 移除临时容器
	//RemoveContainer(d.Info.Id)
}
func initImageInfo(tag string) *ImageInfo {
	//获取镜像id
	imageId := ImageId()
	info := &ImageInfo{
		Id:         imageId,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		WorkDir:    "/",
	}
	tagSplits := strings.Split(tag, ":")
	info.Name = tagSplits[0]
	if len(tagSplits) == 2 {
		info.Version = tagSplits[1]
	}
	return info
}
func initDockerFile() *DockerFile {
	return &DockerFile{
		// 默认的工作目录
		WorkDir:    "/",
		Env:        []string{},
		Volume:     []string{},
		CMD:        []string{},
		EntryPoint: []string{},
		Expose:     []string{},
	}

}
func (d *DockerFile) from(f string) {
	f = strings.TrimPrefix(f, FROM)
	d.From = strings.Trim(f, " ")
	d.Info = BuildFrom(d.From)
}
func (d *DockerFile) run(r string) {
	r = strings.TrimPrefix(r, RUN)
	r, b := isArrayType(r)
	cmd := &CommandArray{
		WorkDir: d.WorkDir,
	}
	if b {
		cmd.Cmds = parseArray(r)
	} else {
		cmd.Cmds = []string{"sh", "-c", strings.Join(parseCommandLine(r), " ")}
	}
	BuildRun(d, cmd)
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
	//最后一个是要拷贝到的地方
	target := list[len(list)-1]
	cpTarget := ""
	//绝对路径
	if strings.HasPrefix(target, "/") {
		cpTarget = path.Join(d.Info.BaseUrl, "merged", target)
	} else {
		//相对路径，此时要拼接workdir
		cpTarget = path.Join(d.Info.BaseUrl, "merged", d.WorkDir, target)
	}
	pwd, _ := os.Getwd()
	for i := 0; i < len(list)-1; i++ {
		// 自动解压归档文件
		if path.Ext(list[i]) == ".tar" {
			UnTar(path.Join(pwd, list[i]), cpTarget)
		} else {
			Copy(path.Join(pwd, list[i]), cpTarget)
		}
	}
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
	//最后一个是要拷贝到的地方
	target := list[len(list)-1]
	cpTarget := ""
	//绝对路径
	if strings.HasPrefix(target, "/") {
		cpTarget = path.Join(d.Info.BaseUrl, "merged", target)
	} else {
		//相对路径，此时要拼接workdir
		cpTarget = path.Join(d.Info.BaseUrl, "merged", d.WorkDir, target)
	}
	pwd, _ := os.Getwd()
	for i := 0; i < len(list)-1; i++ {
		// 拷贝文件
		Copy(path.Join(pwd, list[i]), cpTarget)
	}
}
func (d *DockerFile) expose(e string) {
	e = strings.TrimPrefix(e, EXPOSE)
	// 端口列表
	ports := parseCommandLine(e)
	d.Expose = ports
}
func (d *DockerFile) env(e string) {
	e = strings.TrimPrefix(e, ENV)
	//去掉开头的空格
	e = strings.Trim(e, " ")
	d.Env = append(d.Env, parseEnv(e)...)
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
	d.Volume = parseCommandLine(v)
}
func (d *DockerFile) workDir(w string) {
	w = strings.TrimPrefix(w, WORKDIR)
	d.WorkDir = strings.Trim(path.Clean(w), " ")
}

func (d *DockerFile) copy2ImageInfo(info *ImageInfo) {
	info.WorkDir = d.WorkDir
	info.From = d.From
	info.Env = d.Env
	info.Volume = d.Volume
	info.CMD = d.CMD
	info.EntryPoint = d.EntryPoint
	info.EntryPointShellType = d.EntryPointShellType
	info.CMDShellType = d.CMDShellType
	info.Expose = d.Expose
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
	var array []string
	_ = json.Unmarshal([]byte(s), &array) // 忽略错误，可根据需要处理
	return array
}

// 解析 非数组类型
func parseCommandLine(s string) []string {
	le := len(s)
	cmd := make([]string, 0, 8) // 预分配容量

	i := 0
	for i < le {
		// 跳过前导空格
		for i < le && s[i] == ' ' {
			i++
		}
		if i >= le {
			break
		}

		// 如果参数是引号包裹的
		if s[i] == '"' {
			i++ // 跳过开头的引号
			start := i
			for i < le {
				if s[i] == '"' && (i == start || s[i-1] != '\\') {
					cmd = append(cmd, s[start:i])
					i++ // 跳过结尾引号
					break
				}
				i++
			}
		} else {
			// 普通参数
			start := i
			for i < le && s[i] != ' ' && s[i] != '"' {
				i++
			}
			cmd = append(cmd, s[start:i])
		}
	}
	return cmd
}

func parseEnv(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// 检测是否是单环境变量格式
	firstSpace := strings.IndexByte(s, ' ')
	firstEq := strings.IndexByte(s, '=')
	if firstSpace != -1 && (firstEq == -1 || firstSpace < firstEq) {
		// 单环境变量
		key := s[:firstSpace]
		val := strings.TrimSpace(s[firstSpace+1:])
		return []string{key + "=" + val}
	}

	// 多环境变量解析
	var env []string
	i := 0
	for i < len(s) {
		// 解析 key
		startKey := i
		for i < len(s) && s[i] != '=' {
			i++
		}
		if i >= len(s) {
			break
		}
		key := strings.TrimSpace(s[startKey:i]) + "="
		i++ // 跳过 '='

		// 解析 value
		var valBuilder strings.Builder
		if i < len(s) && s[i] == '"' {
			// 引号包裹的值
			i++
			for i < len(s) {
				if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
					i++
					break
				}
				valBuilder.WriteByte(s[i])
				i++
			}
		} else {
			// 普通值
			for i < len(s) && s[i] != ' ' {
				// 处理转义空格
				if s[i] == '\\' && i+1 < len(s) && s[i+1] == ' ' {
					valBuilder.WriteByte(' ')
					i += 2
					continue
				}
				valBuilder.WriteByte(s[i])
				i++
			}
		}

		env = append(env, key+valBuilder.String())

		// 跳过空格，继续解析下一个
		for i < len(s) && s[i] == ' ' {
			i++
		}
	}
	return env
}

func ResolveImageId(idOrName string, justName bool) string {
	infoList := GetImageInfoList()
	// 先从名称匹配
	var matched []string
	for _, info := range infoList {
		infoName := info.Name
		if info.Version != "" {
			infoName += ":" + info.Version
		}
		if infoName == idOrName {
			matched = append(matched, info.Id)
		}
	}
	if len(matched) > 1 {
		return ""
	}
	if len(matched) == 1 {
		return matched[0]
	}
	if !justName {
		for _, info := range infoList {
			if strings.HasPrefix(info.Id, idOrName) {
				matched = append(matched, info.Id)
			}
		}
		if (len(matched)) > 1 {
			return ""
		}
		if len(matched) == 1 {
			return matched[0]
		}
	}
	return ""
}
