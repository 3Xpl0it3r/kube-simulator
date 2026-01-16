package manager

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type CNIPlugin struct {
	sync.Mutex
	// IP 前缀,例如192.168.0.
	ipPrefix string
	// 子网掩码位数，例如24
	subnetMaskBits string
	// 子网包含的最大IP个数
	maxIPCount uint16
	// 最后分配的IP索引
	lastAssignedIPIndex uint16
	// 已分配的IP地址集合,只存放IP最后一位
	usedIPs map[uint16]struct{}
	// 可用的IP地址集合,只存放IP最后一位
	freeIPs map[uint16]struct{}
}

func NewCNIPlugin(networkCIDR string) (*CNIPlugin, error) {
	_, ipNet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return nil, fmt.Errorf("networkcidr %s is invalid: %v", networkCIDR, err)
	}

	if ipNet.IP.To4() == nil {
		return nil, fmt.Errorf("networkcidr %s must be IPv4", networkCIDR)
	}

	ipParts := strings.Split(ipNet.IP.String(), ".")
	if len(ipParts) != 4 {
		return nil, fmt.Errorf("networkcidr %s has invalid IP format", networkCIDR)
	}

	mask, _ := ipNet.Mask.Size()

	// 计算主机位数和最大IP个数，防止溢出
	hostBits := 32 - mask
	var maxIPs uint16
	if hostBits >= 16 {
		maxIPs = 65535 // uint16最大值
	} else {
		maxIPs = uint16(1 << hostBits)
	}

	plugin := &CNIPlugin{
		maxIPCount:          maxIPs,
		lastAssignedIPIndex: 0,
		usedIPs:             make(map[uint16]struct{}),
		freeIPs:             make(map[uint16]struct{}),
		ipPrefix:            strings.Join(ipParts[:3], "."),
		subnetMaskBits:      strconv.Itoa(mask),
	}

	return plugin, nil
}

// AllocatePodIp [#TODO](should add some comments)
func (p *CNIPlugin) AllocatePodIp() (string, error) {
	p.Lock()
	defer p.Unlock()

	freeCnt := len(p.freeIPs)
	if freeCnt != 0 {
		var allocated uint16
		for k, _ := range p.freeIPs {
			allocated = k
			break
		}
		delete(p.freeIPs, allocated)
		p.usedIPs[allocated] = struct{}{}
		return fmt.Sprintf("%s.%d", p.ipPrefix, allocated), nil
	}

	if p.lastAssignedIPIndex >= p.maxIPCount || p.lastAssignedIPIndex >= 255 {
		return "", errors.New("no more available IP addresses on the node.")
	}

	p.lastAssignedIPIndex += 1
	p.usedIPs[p.lastAssignedIPIndex] = struct{}{}

	return fmt.Sprintf("%s.%d", p.ipPrefix, p.lastAssignedIPIndex), nil
}

// DealloctePodIp [#TODO](should add some comments)
func (p *CNIPlugin) DealloctePodIp(ip string) error {
	p.Lock()
	defer p.Unlock()

	// 更安全的IP解析：先检查IP是否属于当前子网
	if !strings.HasPrefix(ip, p.ipPrefix+".") {
		return fmt.Errorf("IP %s does not belong to network %s", ip, p.ipPrefix)
	}

	// 提取主机部分
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IP format: %s", ip)
	}

	hostAdr, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("invalid host address in IP %s: %v", ip, err)
	}

	// 检查主机地址范围
	if hostAdr < 0 || hostAdr > 255 {
		return fmt.Errorf("host address %d out of range in IP %s", hostAdr, ip)
	}

	delete(p.usedIPs, uint16(hostAdr))
	p.freeIPs[uint16(hostAdr)] = struct{}{}
	return nil
}
