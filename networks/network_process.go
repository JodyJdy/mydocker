package networks

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"path"
)

func test() {
	//netlink.LinkAdd("a")
	//netns.Set(nil)
	//net.IPNet{}
	//ipAddllocator.
	//
}

// Network 存储网络的信息
type Network struct {
	// 网络名
	Name string
	// 网络地址段
	IpRange *net.IPNet
	// 网络驱动名
	Driver string
}

// EndPoint 用于连接容器与网络
type EndPoint struct {
	ID string `json:"id"`
	// 使用 veth 连接网络
	Device netlink.Veth `json:"dev"`
	// ip地址
	IPAddress net.IP `json:"ip"`
	// mac地址
	MacAddress net.HardwareAddr `json:"mac"`
	// 所属的网络
	Network *Network
	// 端口映射，
	PortMapping []string
}

// 持久化一个网络
func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer nwFile.Close()
	// 序列化为json，保存
	nwJson, err := json.Marshal(nw)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = nwFile.Write(nwJson)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// 删除网络
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

// 从文件中加载网络
func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		return err
	}
	info, _ := nwConfigFile.Stat()
	nwJson := make([]byte, info.Size())
	_, err = nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nwJson, nw)
	if err != nil {
		fmt.Printf("加载网络信息失败:%v", err)
		return err
	}
	return nil
}
