package containers

import (
	"fmt"
	"os"
	"os/exec"
)

// NewWorkSpace 返回挂载后的merged目录
func NewWorkSpace(containerBaseUrl string) string {
	createLowerDir(containerBaseUrl)
	createUpperDir(containerBaseUrl)
	createWorkDir(containerBaseUrl)
	return createMergedDir(containerBaseUrl)
}

// 创建只读层
func createLowerDir(containerBaseUrl string) error {
	lowerDir := containerBaseUrl + "lower/"
	busyboxTarUrl := containerBaseUrl + "busybox.tar"

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
			fmt.Print("Untar dir %s error %v", lowerDir, err)
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
func DeleteWorkSpace(containerBaseUrl string) {
	DeleteMountPoint(containerBaseUrl)
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
func DeleteLowerDir(containerBaseUrl string) {

}
func DeleteWorkDir(containerBaseUrl string) {

}

func DeleteUpperDir(containerBaseUrl string) {

}
