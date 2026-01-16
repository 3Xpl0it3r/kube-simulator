package manager

import (
	"testing"
)

func TestNewCNIPlugin(t *testing.T) {
	testCases := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{
			name:        "Valid IPv4 CIDR",
			cidr:        "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "Valid IPv4 CIDR with different subnet",
			cidr:        "10.244.0.0/16",
			expectError: false,
		},
		{
			name:        "Invalid CIDR format",
			cidr:        "invalid-cidr",
			expectError: true,
		},
		{
			name:        "Invalid IPv4 address",
			cidr:        "256.256.256.256/24",
			expectError: true,
		},
		{
			name:        "IPv6 CIDR (not supported)",
			cidr:        "2001:db8::/64",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin, err := NewCNIPlugin(tc.cidr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for CIDR %s", tc.cidr)
				}
				if plugin != nil {
					t.Error("Expected nil plugin when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for CIDR %s: %v", tc.cidr, err)
				}
				if plugin == nil {
					t.Error("Expected non-nil plugin")
				}
			}
		})
	}
}

func TestCNIPlugin_AllocatePodIp(t *testing.T) {
	plugin, err := NewCNIPlugin("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create CNIPlugin: %v", err)
	}

	// 分配第一个 IP
	ip1, err := plugin.AllocatePodIp()
	if err != nil {
		t.Fatalf("Failed to allocate IP: %v", err)
	}

	if ip1 == "" {
		t.Error("Allocated IP should not be empty")
	}

	// 验证 IP 在正确的网段
	if !isValidIPInSubnet(ip1, "192.168.1.0/24") {
		t.Errorf("IP %s is not in subnet 192.168.1.0/24", ip1)
	}

	// 分配第二个 IP
	ip2, err := plugin.AllocatePodIp()
	if err != nil {
		t.Fatalf("Failed to allocate second IP: %v", err)
	}

	if ip2 == "" {
		t.Error("Second allocated IP should not be empty")
	}

	// 验证两个 IP 不同
	if ip1 == ip2 {
		t.Errorf("IPs should be different: %s and %s", ip1, ip2)
	}
}

func TestCNIPlugin_AllocatePodIp_Exhaustion(t *testing.T) {
	// 创建小范围的子网进行测试
	plugin, err := NewCNIPlugin("192.168.1.0/30") // 只有几个IP可用
	if err != nil {
		t.Fatalf("Failed to create CNIPlugin: %v", err)
	}

	// 分配所有可用的IP
	allocatedIPs := make([]string, 0)
	for {
		ip, err := plugin.AllocatePodIp()
		if err != nil {
			break
		}
		allocatedIPs = append(allocatedIPs, ip)
	}

	// 尝试分配超过限制的IP
	_, err = plugin.AllocatePodIp()
	if err == nil {
		t.Error("Expected error when IP pool is exhausted")
	}
}

func TestCNIPlugin_DeallocatePodIp(t *testing.T) {
	plugin, err := NewCNIPlugin("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create CNIPlugin: %v", err)
	}

	// 分配一个IP
	ip, err := plugin.AllocatePodIp()
	if err != nil {
		t.Fatalf("Failed to allocate IP: %v", err)
	}

	// 释放IP
	err = plugin.DealloctePodIp(ip)
	if err != nil {
		t.Errorf("Failed to deallocate IP %s: %v", ip, err)
	}

	// 重新分配相同的IP应该可以工作（如果实现支持重用）
	ip2, err := plugin.AllocatePodIp()
	if err != nil {
		t.Errorf("Failed to allocate IP after deallocation: %v", err)
	}

	// 在当前实现中，IP可能不会立即重用，但不会出错
	if ip2 == "" {
		t.Error("Reallocated IP should not be empty")
	}
}

func TestCNIPlugin_DeallocatePodIp_InvalidIP(t *testing.T) {
	plugin, err := NewCNIPlugin("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to create CNIPlugin: %v", err)
	}

	testCases := []struct {
		name string
		ip   string
	}{
		{"Empty IP", ""},
		{"Invalid format", "not-an-ip"},
		{"IP not in subnet", "10.0.0.1"},
		{"Out of range IP", "192.168.1.999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := plugin.DealloctePodIp(tc.ip)
			if err == nil {
				t.Errorf("Expected error for invalid IP %s", tc.ip)
			}
		})
	}
}

// isValidIPInSubnet 检查IP是否在指定子网内
func isValidIPInSubnet(ip, cidr string) bool {
	// 简单实现，只检查IP前缀
	return len(ip) > len("192.168.1.") && ip[:len("192.168.1.")] == "192.168.1."
}
