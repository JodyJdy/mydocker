package portmapping

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

// 全局锁
var lock sync.Mutex

// 全局映射信息
var portMapping *PortMapping

// PortMapping 记录端口映射的所有信息
type PortMapping struct {
	// 记录端口以及端口映射地址
	TargetWithPort map[int][]string
	// 记录端口是否映射
	MappingStatus map[int]bool
	// 映射多个地址时，记录负载均衡的信息
	LoadBalancer map[int]int
}

func (p *PortMapping) addPortMapping(port int, addr string) {
	log.Printf("添加端口映射: %d:%v\n", port, addr)
	status := p.MappingStatus[port]
	lock.Lock()
	p.TargetWithPort[port] = append(p.TargetWithPort[port], addr)
	p.MappingStatus[port] = true
	lock.Unlock()

	// 第一次添加该端口的映射
	if !status {
		go listen(port)
	}
}
func (p *PortMapping) removePortMapping(port int, addr string) {
	lock.Lock()
	s := p.TargetWithPort[port]
	var result []string
	for _, v := range s {
		// 去掉addr
		if v != addr {
			result = append(result, v)
		}
	}
	p.TargetWithPort[port] = result
	if len(result) == 0 {
		p.MappingStatus[port] = false
	}
	lock.Unlock()
}
func (p *PortMapping) removePort(port int) {
	lock.Lock()
	p.TargetWithPort[port] = []string{}
	p.MappingStatus[port] = false
	lock.Unlock()
}

// 负载均衡
func (p *PortMapping) loadBalancer(port int) string {
	// 端口映射关闭，不再接收连接
	if !p.MappingStatus[port] {
		return ""
	}
	lock.Lock()
	defer lock.Unlock()
	p.LoadBalancer[port] = (p.LoadBalancer[port] + 1) % len(p.TargetWithPort[port])
	return p.TargetWithPort[port][p.LoadBalancer[port]]
}
func listen(port int) {
	fromaddr := fmt.Sprintf("0.0.0.0:%d", port)
	fromlistener, err := net.Listen("tcp", fromaddr)
	if fromlistener == nil {
		return
	}
	defer fromlistener.Close()
	if err != nil {
		log.Printf("监听端口 %s, 失败: %s\n", fromaddr, err.Error())
		return
	}
	// 进入循环
	for portMapping.MappingStatus[port] {
		// 接收连接
		fromcon, err := fromlistener.Accept()
		//这边最好也做个协程，防止阻塞
		toAddr := portMapping.loadBalancer(port)
		log.Printf("接收连接  fro:%s to %s\n", fromaddr, toAddr)
		toCon, err := net.Dial("tcp", toAddr)
		if err != nil {
			log.Printf("不能连接到 %s\n", toAddr)
			fromcon.Close()
			continue
		}
		// 拷贝连接信息
		go io.Copy(fromcon, toCon)
		io.Copy(toCon, fromcon)
	}
}

func AddPortMapping(strPort string, addr string) {
	port, _ := strconv.Atoi(strPort)
	// 如果没有执行初始化，执行初始化
	if portMapping == nil {
		portMapping = initPortMapping()
	}
	portMapping.addPortMapping(port, addr)
}
func initPortMapping() *PortMapping {
	return &PortMapping{
		TargetWithPort: make(map[int][]string),
		MappingStatus:  make(map[int]bool),
		LoadBalancer:   make(map[int]int),
	}
}
func RemovePortMapping(strPort string, addr string) {
	port, _ := strconv.Atoi(strPort)
	if portMapping == nil {
		portMapping = initPortMapping()
		return
	}
	portMapping.removePortMapping(port, addr)
}
func RemovePort(strPort string) {
	port, _ := strconv.Atoi(strPort)
	if portMapping == nil {
		portMapping = initPortMapping()
		return
	}
	portMapping.removePort(port)
}

func Main() {
	ch := make(chan int)
	fmt.Println(<-ch)
}
