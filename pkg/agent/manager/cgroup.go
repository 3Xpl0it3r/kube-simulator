package manager

import coreapi "k8s.io/api/core/v1"

// CGrpupManager represent cgroupmanager
type CGrpupManager struct {
	MemTotalInBytes int64
	MemUsedInBytes  int64
	CpuTotal        int64
	CpuUsed         int64
	DiskTotal       int64
	DiskUsage       int32
}

func NewCGroupManager(node *coreapi.Node) *CGrpupManager {
	return &CGrpupManager{}
}

func (c *CGrpupManager) OnAdd(pod *coreapi.Pod) {
	var memNeed, cpuNeed = podRequestedResources(pod)
	c.MemUsedInBytes += memNeed
	c.CpuUsed += cpuNeed
}

func (c *CGrpupManager) OnUpdate(pod *coreapi.Pod) {
	var memNeed, cpuNeed = podRequestedResources(pod)
	c.MemUsedInBytes += memNeed
	c.CpuUsed += cpuNeed

}

// OnAdd [#TODO](should add some comments)
func (c *CGrpupManager) OnDelete(pod *coreapi.Pod) {
	memNeed, cpuNeed := podRequestedResources(pod)
	c.MemUsedInBytes -= memNeed
	c.CpuUsed -= cpuNeed
}

// Merge [#TODO](should add some comments)
func (c *CGrpupManager) Merge(newObj *CGrpupManager) {
	c.MemTotalInBytes = newObj.MemTotalInBytes
	c.CpuTotal = newObj.CpuTotal
	c.DiskTotal = newObj.DiskTotal
}

// Allocateable [#TODO](should add some comments)
func (c *CGrpupManager) HasSufficientResourcesForWorload(pod *coreapi.Pod) bool {
	memNeed, cpuNeed := podRequestedResources(pod)
	return (c.MemTotalInBytes > c.MemUsedInBytes+memNeed) && (c.CpuTotal > c.CpuUsed+cpuNeed)
}

func podRequestedResources(pod *coreapi.Pod) (mem, cpu int64) {
	for _, container := range pod.Spec.Containers {
		mem += container.Resources.Requests.Memory().Value()
		cpu += container.Resources.Requests.Cpu().Value()
	}
	return
}
