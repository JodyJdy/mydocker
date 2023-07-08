package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// SetUpMount 初始化挂载点
func SetUpMount() {
	//获取工作目录
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前工作目录失t"+
			"：%v \n", err)
	}
	//挂载root目录
	err = pivotRoot(pwd)
	if err != nil {
		_ = fmt.Errorf("挂载root目录失败: %v \n ", err)
		return
	}

	//设置默认挂载参数 MS_NOEXEC本文将系统不允许运行其他程序 MS_NOEXEC 运行程序的时候
	// 不允许 set-user-id,set-group-id
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载 proc目录，使 容器有独立的proc目录
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		_ = fmt.Errorf("挂载 /proc 目录 失败: %v \n", err)
		return
	}
	//挂载 /dev 目录
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		fmt.Printf("挂载 /dev 目录 失败: %v \n", err)
		return
	}

}

func pivotRoot(containerRoot string) error {
	// 使用 mount --bind foo foo 方式 将 containerRoot重新挂载了一次
	//使得 cotainerRoot的文件系统和 宿主机的文件系统不同
	//这是 pivot_root的必须要求
	if err := syscall.Mount(containerRoot, containerRoot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		fmt.Printf("mount rootfs to itself error: %v", err)
		return err
	}
	// 创建 cotainerRoot/.pivot_root 存储 old_root
	pivotDir := filepath.Join(containerRoot, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在cotainerRoot/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	// 根目录已经被替换
	if err := syscall.PivotRoot(containerRoot, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	// 卸载掉 containerRoot/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹，如果不删除的话，就可以在容器里面访问到宿主机的根目录
	return os.Remove(pivotDir)
}
