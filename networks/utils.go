package networks

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

func createBridgeInterface(bridgeName string) error {
	// 检查是否有同名的 网络设备
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	// 创建一个netlink类型的对象
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	// 使用 netlink对象创建 bridge
	br := &netlink.Bridge{LinkAttrs: la}
	// 相当于 ip link add xxx
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge 创建失败 %s: %v", bridgeName, err)
	}
	return nil
}

// SetInterfaceUP 启动bridge 设备
func SetInterfaceUP(interfaceName string) error {
	// 找到 bridge设备
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("设备找不到 [ %s ]: %v", iface.Attrs().Name, err)
	}
	// 启动 link
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("启动bridge %s失败: %v", interfaceName, err)
	}
	return nil
}

// SetInterfaceIP 设置bridge ip地址
func SetInterfaceIP(name string, rawIP string) error {
	retries := 2
	// 找到需要设置的bridge
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		fmt.Printf("bridge link 不存在[ %s ]...\n", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	// 返回的ipnet中 包含了网段的信息，也包含了原始的ip地址
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	//给 网路接口配置地址 相当于 ip addr add xxx
	//由于ipNet中包含了网段的信息，还会配置路由表，将 xx.xx.0.0/16  转发到这个网络接口中
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0, Broadcast: nil}
	return netlink.AddrAdd(iface, addr)
}

// SetContainerNetNs 设置指定进程的网络命名空间
func SetContainerNetNs(pid string) {
	// 访问容器进程pid目录下的net文件
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("获取网络命名空间失败, %v\n", err)
	}
	nsFD := f.Fd()
	// 锁定线程 go是多线程，进入ns时需要锁定线程
	runtime.LockOSThread()
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		fmt.Printf("设置网络命名空间失败, %v\n", err)
	}
}
