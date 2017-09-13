package subsystems

// ResourceConfig is configs of resources.
type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

// SubSystem is cgourp subsystem, and path is a cgroup.
type SubSystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

// Initialize a instance using subsystems.
var (
	SubSystemsIns = []SubSystem{
		// &CpusetSubSystem{},
		&MemorySubSystem{},
		// &CpuSubSystem{},
	}
)
