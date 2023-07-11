package networks

const IpamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

const DefaultNetworkPath = "/var/run/mydocker/network/network/"

const DefaultDriver = "bridge"

// 驱动
var drivers = map[string]NetworkDriver{}

// 网络
var networks = map[string]*Network{}

// NONE 网络模式,默认的，什么都不做
var NONE = "none"

// HOST 和主机共享网络
var HOST = "host"

// CONTAINER 与容器共享网络
var CONTAINER = "container"
