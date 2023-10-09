package networks

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

// IpAllocatorManager Ip地址分配管理器
type IpAllocatorManager struct {
	// 分配文件存放位置
	SubnetAllocatorPath string
	// key是网段 values 是网段中的 位图算法，记录网段中的ip分配情况 string中的一个字符表示一个状态位
	Subnets *map[string]string
}

// 从文件中加载信息
func (ipam *IpAllocatorManager) load() error {
	fileInfo, err := os.Stat(ipam.SubnetAllocatorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, fileInfo.Size())
	_, err = subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(subnetJson, ipam.Subnets)
	if err != nil {
		log.Printf("加载ipam失败 %v", err)
		return err
	}
	return nil
}
func (ipam *IpAllocatorManager) dump() error {
	// 如果目录不能存在就创建
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	//O_TRUNC  存在内容就是清空;O_CREATE 不存在就创建；O_WRONLY 只写
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	//序列化为json
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	// 写到文件
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}
	return nil
}

// Allocate 分配网络的ip
func (ipam *IpAllocatorManager) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}

	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		fmt.Printf("加载ipam 失败%v\n", err)
	}
	// 获取网络
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// 如果网络为 127.0.0.1/8,掩码是255.0.0.0   one=8,size=32, one是前面1的个数，size是总个数
	one, size := subnet.Mask.Size()
	// 如果不存在，创建，   2^ (size-one) 就是当前网络中，可以分配的所有ip地址，初始为全0
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	for c := range (*ipam.Subnets)[subnet.String()] {
		// 遍历到的第一个0，就是网络
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// 字符串转成byte 数组
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			//设置为1，表示这个网络被分配
			ipalloc[c] = '1'
			// 字符串不能更改， 还原回去
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			// 计算当前分配的ip
			ip = subnet.IP
			// 如果当序号是x，uint8表示取低8位，在原有网络地址的基础上，加上 分配的网络号，那么第一个字节加的内容是 uint8(x>>24),
			// 第二个字节是  uint8(x>>16), 第三个 uint8(x>>8), 第四个 uint8(x)

			// 如果是 172.16.0.0/16, c = 0， 也就是第一个地址
			// 那么分配后的ip是 172.16.0.1
			ip[0] += uint8(c >> 24)
			ip[1] += uint8(c >> 16)
			ip[2] += uint8(c >> 8)
			ip[3] += uint8(c)
			//地址分配从1开始，默认要加1
			ip[3] += 1
			break
		}
	}
	// 记录更改
	ipam.dump()
	return
}

// Release 释放地址
func (ipam *IpAllocatorManager) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Printf("加载ipam 失败%v\n", err)
	}

	c := 0
	// 转成四个字节
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	// 对应的byte位设置为0，表示网络释放
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}
