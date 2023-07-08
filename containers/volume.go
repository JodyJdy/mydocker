package containers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// NewWorkSpace 返回挂载后的merged目录
func NewWorkSpace(info *ContainerInfo, volume string, imageId string) string {
	lowDir := getLowerDir(imageId)
	createUpperDir(info.BaseUrl)
	createWorkDir(info.BaseUrl)
	mergedDir := createMergedDir(info.BaseUrl, lowDir)
	//创建卷的挂载
	CreateVolume(info, mergedDir, volume, imageId)
	return mergedDir
}

// 获取只读层 目录
func getLowerDir(image string) string {
	var lowDirs []string
	lowDirs = append(lowDirs, fmt.Sprintf(ImageLayerLocation, image))
	info, err := GetImageInfo(image)
	if err != nil {
		fmt.Println("镜像不存在")
	}
	fmt.Println(info)
	for {
		//按层查找
		if info.From != "" {
			from := info.From
			info, err = GetImageInfo(ResolveImageId(info.From, false))
			lowDirs = append(lowDirs, fmt.Sprintf(ImageLayerLocation, from))
		} else {
			break
		}
	}
	fmt.Println("lowdir 是:")
	fmt.Println(lowDirs)
	return strings.Join(lowDirs, ":")
}
func createUpperDir(containerBaseUrl string) {
	upperDir := path.Join(containerBaseUrl, "upper")
	if err := os.MkdirAll(upperDir, 0777); err != nil {
		fmt.Printf("Mkdir upper layer dir %s error. %v", upperDir, err)
	}

}
func createWorkDir(containerBaseUrl string) {
	workDir := path.Join(containerBaseUrl, "work")
	if err := os.MkdirAll(workDir, 0777); err != nil {
		fmt.Printf("Mkdir work layer dir %s error. %v", workDir, err)
	}
}

//	mount -t overlay  overlay  \
//	             -olowerdir=/lower,upperdir=/upper,workdir=/work  /merged
func createMergedDir(containerBaseUrl string, lowDir string) string {
	mergedDir := path.Join(containerBaseUrl, "merged")
	if err := os.MkdirAll(mergedDir, 0777); err != nil {
		fmt.Printf("Mkdir merged layer dir %s error. %v", mergedDir, err)
	}
	// 处理卷挂载
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowDir, containerBaseUrl+"upper/", containerBaseUrl+"work/")
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "none", mergedDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v \n  ", err)
	}
	return mergedDir
}
func CreateVolume(info *ContainerInfo, mergedDir string, volume string, imageId string) {
	if volume != "" {
		volumeUrl := strings.Split(volume, ":")
		// 宿主机路径
		hostUrl := volumeUrl[0]
		//容器内路径
		containerPath := volumeUrl[1]
		MountVolume(info, hostUrl, containerPath, mergedDir, false)
	}
	// 创建匿名卷
	imageInfo, err := GetImageInfo(imageId)
	if err != nil {
		fmt.Printf("获取镜像失败:%v", err)
	}
	for _, v := range imageInfo.Volume {
		// 生成卷id
		vId := VolumeId()
		volumeUrl := fmt.Sprintf(VolumeInfoLocation, vId)
		MountVolume(info, volumeUrl, v, mergedDir, true)
	}

}

// MountVolume hostPath 挂载卷的位置 containerPath被挂载的路径（容器中）， mergedPath 容器宿主机工作路径
func MountVolume(info *ContainerInfo, hostPath string, containerPath string, mergedPath string, anonymous bool) {
	containerPathInHost := path.Join(mergedPath, containerPath)
	exist, _ := PathExists(hostPath)
	//创建路径
	if !exist {
		err := os.MkdirAll(hostPath, 0777)
		if err != nil {
			fmt.Printf("创建宿主机目录: %s，失败: %v \n", hostPath, err)
			return
		}
	}
	exist, _ = PathExists(containerPathInHost)
	if !exist {
		//创建目录
		err := os.MkdirAll(containerPathInHost, 0777)
		if err != nil {
			fmt.Printf("创建容器中卷的目录:%s,失败: %v \n", containerPathInHost, err)
			return
		}
	}
	//挂载
	cmd := exec.Command("mount", "--bind", hostPath, containerPathInHost)
	// 打印命令的输入输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("挂载卷失败： %v", err)
	}
	// 添加卷
	info.Volume = append(info.Volume, VolumeInfo{
		HostVolumePath:      hostPath,
		ContainerPathInHost: containerPathInHost,
		ContainerPath:       containerPath,
		Anonymous:           anonymous,
	})

}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteWorkSpace 当删除容器时，会删除相关的目录
func DeleteWorkSpace(info *ContainerInfo) {
	DeleteVolumeMount(info)
	DeleteOverlayMountPoint(info.BaseUrl)
}

func DeleteOverlayMountPoint(containerBaseUrl string) {
	mergedDir := path.Join(containerBaseUrl, "merged")
	umount(mergedDir)
}
func DeleteVolumeMount(info *ContainerInfo) {
	for _, v := range info.Volume {
		umount(v.ContainerPathInHost)
	}
}
func umount(mountedPath string) {
	cmd := exec.Command("umount", mountedPath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("unmount mount point  %s failed %v ", mountedPath, err)
		return
	}
}
