package cgroups

import "github.com/yungkcx/mydocker/cgroups/subsystems"

import "github.com/Sirupsen/logrus"

// CgroupManager is for manage resources.
type CgroupManager struct {
	// Path of the cgroup in hierarchy.
	Path     string
	Resource *subsystems.ResourceConfig
}

// NewCgroupManager returns a new CgroupManager.
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply add the process to subsystems of the cgroup.
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubSystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set sets resource limitation of the cgroup.
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubSystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

// Destroy removes the cgroup.
func (c *CgroupManager) Destroy() error {
	logrus.Infoln("This is a test info")
	for _, subSysIns := range subsystems.SubSystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("remove cgroup fail %v", err)
		}
	}
	return nil
}