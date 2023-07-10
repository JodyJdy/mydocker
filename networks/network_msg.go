package networks

const IpamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

const DefaultNetworkPath = "/var/run/mydocker/network/network/"

// 驱动
var drivers = map[string]NetworkDriver{}

// 网络
var networks = map[string]*Network{}
