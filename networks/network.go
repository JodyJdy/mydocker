package networks

import (
	"containers"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var ipAllocatorManager = &IpAllocatorManager{
	SubnetAllocatorPath: IpamDefaultAllocatorPath,
}

// Init 初始化时加载网络
func Init() error {
	// 创建网桥
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(DefaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(DefaultNetworkPath, 0644)
		} else {
			return err
		}
	}
	filepath.Walk(DefaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		if err := nw.load(nwPath); err != nil {
			fmt.Printf("加载网络: %s失败\n", err)
		}
		networks[nwName] = nw
		return nil
	})
	return nil
}
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		return
	}
}

// CreateNetwork 创建网络
func CreateNetwork(driver, subnet, name string) error {
	// 解析cidr网络  127.0.0.1/8
	_, cidr, _ := net.ParseCIDR(subnet)
	// 分配ip
	ip, err := ipAllocatorManager.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	// 调用驱动创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	// 存储网络
	return nw.dump(DefaultNetworkPath)
}

// Connect 容器连接到网络
func Connect(networkName string, cinfo *containers.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		fmt.Printf("网络不存在: %s\n", networkName)
		return nil
	}
	// 分配容器IP地址
	ip, err := ipAllocatorManager.Allocate(network.IpRange)
	if err != nil {
		return err
	}
	// 创建网络端点
	ep := &EndPoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的namespace配置容器网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	// 配置容器到网络中的端口映射
	return configPortMapping(ep)
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		fmt.Printf("网络不存在: %s\n", networks)
		return nil
	}
	// 删除网络
	if err := ipAllocatorManager.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("删除网络网关地址失败: %s", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("驱动删除网络失败 %s", err)
	}
	return nw.remove(DefaultNetworkPath)
}

func configEndpointIpAddressAndRoute(ep *EndPoint, cinfo *containers.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// 进入到容器的 netns
	// 退出时，还原 netns
	defer enterContainerNetns(&peerLink, cinfo)()

	// 当前处于容器内的netns， 配置 veth在容器端的 ip地址
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	// 设置 veth ip地址
	if err = SetInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}
	// 启动 veth
	if err = SetInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	// 默认的 127.0.0.1 是关闭的， 打开
	if err = SetInterfaceUP("lo"); err != nil {
		return err
	}
	// 设置容器的外部请求都从 veth 访问
	// 例如 veth 就是最后 Use IFace
	//Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
	//0.0.0.0         192.168.0.1     0.0.0.0         UG    100    0        0 eth0
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		// 网关地址
		Gw:  ep.Network.IpRange.IP,
		Dst: cidr,
	}
	// 添加路由规则
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

func configPortMapping(ep *EndPoint) error {
	// 配置访问外部网络
	// 配置端口映射
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			fmt.Printf("端口映射格式错误, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("iptalbes输出 %v", output)
			continue
		}
	}
	return nil
}

// 进入容器netns
func enterContainerNetns(enLink *netlink.Link, cinfo *containers.ContainerInfo) func() {
	// 访问容器进程pid目录下的net文件
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("error get container net namespace, %v\n", err)
	}

	nsFD := f.Fd()
	// 锁定线程 go是多线程，进入ns时需要锁定线程
	runtime.LockOSThread()
	// 修改veth peer 另外一端移到容器的namespace中，将veth的另一端连接到容器
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		fmt.Printf("error set link netns , %v\n", err)
	}

	// 获取当前的网络namespace, 用于后面还原
	origns, err := netns.Get()
	if err != nil {
		fmt.Printf("error get current netns, %v\n", err)
	}
	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		fmt.Printf("error set netns, %v\n", err)
	}
	return func() {
		// 还原netns
		netns.Set(origns)
		origns.Close()
		// 解锁线程
		runtime.UnlockOSThread()
		f.Close()
	}
}
