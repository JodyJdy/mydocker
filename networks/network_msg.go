package networks

const (
	IpamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"
	DefaultNetworkPath       = "/var/run/mydocker/network/network/"
	DefaultDriver            = "bridge"
)

// 驱动
var drivers = map[string]NetworkDriver{}

// 网络
var networks = map[string]*Network{}

const (
	// NONE 网络模式,默认的，什么都不做
	NONE = "none"
	// HOST 和主机共享网络
	HOST = "host"
	// CONTAINER 与容器共享网络
	CONTAINER = "container"
)
