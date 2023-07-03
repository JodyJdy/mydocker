package cgroups

// ResourceConfig 传递资源限制的结构体 内存限制，cpu时间权重，cpu核数
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	// Name 返回子系统的名称,例如 cpu,memory
	Name() string
	// Set 设置cgroup中subsystem的资源限制
	Set(path string, res *ResourceConfig) error
	// Apply 将进程添加到cgroup中
	Apply(path string, pid int) error
	// Remove 移除某个cgroup
	Remove(path string) error
}

// SubsystemIns 创建不同的子系统，用于方法调用
var (
	SubsystemIns = []Subsystem{
		&CpuSetSubsystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
