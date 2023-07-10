package networks

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

// NetworkDriver 网络驱动 使用网络驱动 创建/连接/销毁网络
// 不同的驱动使用不同从策略
type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *EndPoint) error
	Disconnect(network Network, endpoint *EndPoint) error
}

type BridgeNetworkDriver struct{}

func (BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	// 解析ip地址
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	err := d.initBridge(n)
	if err != nil {
		fmt.Printf("初始化bridge 失败:%v", err)
	}
	return n, nil
}
func (d *BridgeNetworkDriver) initBridge(n *Network) error {

	// 创建bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error add bridge： %s, Error: %v", bridgeName, err)
	}
	//设置bridge 地址和路由
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := SetInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("error assigning address: %s on bridge: %s with an error of: %v", gatewayIP, bridgeName, err)
	}
	//启动bridge
	if err := SetInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("error set bridge up: %s, Error: %v", bridgeName, err)
	}

	// 设置 iptables
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
	}

	return nil
}

func (BridgeNetworkDriver) Delete(network Network) error {
	// 删除bridge
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

func (BridgeNetworkDriver) Connect(network *Network, endpoint *EndPoint) error {
	bridgeName := network.Name
	// 获取到bridge设备
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 创建 veth对象
	la := netlink.NewLinkAttrs()
	// 名称使用 endpoint id 前5位， 长度不能太长
	la.Name = endpoint.ID[:5]
	// veth的一端 连接到 bridge
	la.MasterIndex = br.Attrs().Index
	// 创建 veth对象
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		// 容器中的veth名称
		PeerName: "cif-" + endpoint.ID[:5],
	}
	// 添加veth
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	// 启动
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

func (BridgeNetworkDriver) Disconnect(network Network, endpoint *EndPoint) error {
	return nil
}

// 设置iptables 规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// 执行命令 配置 snat 规则
	// 只要是 这个网桥上发出的包 ，都会做源ip地址转换，保证了 容器 访问外部网络的数据包使用宿主机ip
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	// 执行命令或获取输出
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("iptables输出:  %v\n", output)
	}
	return err
}
