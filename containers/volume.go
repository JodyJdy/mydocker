package containers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NewWorkSpace 返回挂载后的merged目录
func NewWorkSpace(containerBaseUrl, volume string) string {
	err := createLowerDir(containerBaseUrl)
	if err != nil {
		_ = fmt.Errorf("创建LowerDir失败: %v", err)
		return ""
	}
	createUpperDir(containerBaseUrl)
	createWorkDir(containerBaseUrl)
	mergedDir := createMergedDir(containerBaseUrl)
	//创建卷的挂载
	CreateVolume(mergedDir, volume)
	return mergedDir
}

// 创建只读层
func createLowerDir(containerBaseUrl string) error {
	lowerDir := containerBaseUrl + "lower/"
	// busybox.tar @TODO 这里用于处理镜像，现在使用的是默认的镜像
	busyboxTarUrl := "/root/busybox.tar"
	exist, err := PathExists(lowerDir)
	if err != nil {
		fmt.Printf("Fail to judge whether dir %s exists. %v \n", lowerDir, err)
		return err
	}
	if !exist {
		if err := os.MkdirAll(lowerDir, 0622); err != nil {
			fmt.Printf("Mkdir %s error %v \n", lowerDir, err)
			return err
		}
		if _, err := exec.Command("tar", "--strip-components", "1", "-xvf", busyboxTarUrl, "-C", lowerDir).CombinedOutput(); err != nil {
			fmt.Printf("Untar dir %s error %v\n", lowerDir, err)
			return err
		}
	}
	return nil
}
func createUpperDir(containerBaseUrl string) {
	upperDir := containerBaseUrl + "upper/"
	if err := os.MkdirAll(upperDir, 0777); err != nil {
		fmt.Printf("Mkdir upper layer dir %s error. %v", upperDir, err)
	}

}
func createWorkDir(containerBaseUrl string) {
	workDir := containerBaseUrl + "work/"
	if err := os.MkdirAll(workDir, 0777); err != nil {
		fmt.Printf("Mkdir work layer dir %s error. %v", workDir, err)
	}
}

//	mount -t overlay  overlay  \
//	             -olowerdir=/lower,upperdir=/upper,workdir=/work  /merged
func createMergedDir(containerBaseUrl string) string {
	mergedDir := containerBaseUrl + "merged/"
	if err := os.MkdirAll(mergedDir, 0777); err != nil {
		fmt.Printf("Mkdir merged layer dir %s error. %v", mergedDir, err)
	}
	// 处理卷挂载
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", containerBaseUrl+"lower/", containerBaseUrl+"upper/", containerBaseUrl+"work/")
	// none是挂载名称， @todo 应该每个容器都有自己的一个名称
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "none", mergedDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v \n  ", err)
	}
	return mergedDir
}
func CreateVolume(mergedDir string, volume string) {
	if volume != "" {
		volumeUrl := strings.Split(volume, ":")
		// 宿主机路径
		hostUrl := volumeUrl[0]
		exist, _ := PathExists(hostUrl)
		//创建路径
		if !exist {
			err := os.Mkdir(hostUrl, 0777)
			if err != nil {
				_ = fmt.Errorf("创建宿主机目录: %s，失败: %v \n", hostUrl, err)
				return
			}
		}
		//容器内路径
		containerUrl := volumeUrl[1]
		// mergedDir开头已经包含了 /，需要去掉多余的
		if containerUrl[0] == '/' {
			containerUrl = containerUrl[1:]
		}
		//容器内路径在 宿主机中的位置
		containerUrlInHost := mergedDir + containerUrl
		exist, _ = PathExists(containerUrlInHost)
		if !exist {
			//创建目录
			err := os.Mkdir(containerUrlInHost, 0777)
			if err != nil {
				_ = fmt.Errorf("创建容器中卷的目录:%s,失败: %v \n", containerUrlInHost, err)
				return
			}
		}
		// 进行挂载
		cmd := exec.Command("mount", "--bind", hostUrl, containerUrlInHost)
		// 打印命令的输入输出
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Errorf("挂载卷失败： %v", err)
		}
	}

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

// DeleteWorkSpace 当删除容器时，会删除相关的目录 @Todo 删除容器中的volume挂载点
func DeleteWorkSpace(containerBaseUrl string, volume string) {
	DeleteVolumeMount(containerBaseUrl, volume)
	DeleteMountPoint(containerBaseUrl)
	upperDir := containerBaseUrl + "upper/"
	workDir := containerBaseUrl + "work/"
	lowerDir := containerBaseUrl + "lower/"
	//删除工作目录所有内容
	er1, er2, er3 := os.RemoveAll(upperDir), os.RemoveAll(workDir), os.RemoveAll(lowerDir)
	if er1 != nil {
		fmt.Printf("RemoveDir %s error %v", upperDir, er1)
		return
	}
	if er2 != nil {
		fmt.Printf("RemoveDir %s error %v", workDir, er2)
		return
	}
	if er3 != nil {
		fmt.Printf("RemoveDir %s error %v", lowerDir, er3)
		return
	}
}

func DeleteMountPoint(containerBaseUrl string) {
	mergedDir := fmt.Sprintf("%smerged/", containerBaseUrl)
	// 卸载merged挂载点
	cmd := exec.Command("umount", mergedDir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v \nj", err)
	}
	if err := os.RemoveAll(mergedDir); err != nil {
		fmt.Printf("RemoveDir %s error %v", mergedDir, err)
	}
}
func DeleteVolumeMount(containerBaseUrl string, volume string) {
	mergedDir := fmt.Sprintf("%smerged/", containerBaseUrl)
	if volume != "" {
		volumeUrl := strings.Split(volume, ":")
		//容器内路径
		containerUrl := volumeUrl[1]
		// mergedDir开头已经包含了 /，需要去掉多余的
		if containerUrl[0] == '/' {
			containerUrl = containerUrl[1:]
		}
		//容器内路径在 宿主机中的位置
		containerUrlInHost := mergedDir + containerUrl

		cmd := exec.Command("umount", containerUrlInHost)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			_ = fmt.Errorf("unmount mount point  %s failed %v ", containerUrlInHost, err)
			return
		}
		err = os.RemoveAll(containerUrlInHost)
		if err != nil {
			_ = fmt.Errorf("remove mount point dir  %s error %v ", containerUrlInHost, err)
			return
		}
	}
}
