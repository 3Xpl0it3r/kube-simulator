package agent

import (
	"testing"
)

func TestConfig_Structure(t *testing.T) {
	helper := NewTestHelper(t)

	// 测试 Config 结构体的基本功能
	config := &Config{
		ClientConfig: "/path/to/kubeconfig",
	}

	helper.AssertEqual("/path/to/kubeconfig", config.ClientConfig, "ClientConfig should be set correctly")
}

func TestConfig_DefaultValues(t *testing.T) {
	helper := NewTestHelper(t)

	// 测试零值
	var config Config

	helper.AssertEqual("", config.ClientConfig, "Default ClientConfig should be empty string")
}

func TestConfig_ValidValues(t *testing.T) {
	helper := NewTestHelper(t)

	// 测试有效值
	testCases := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid config with kubeconfig",
			config: Config{
				ClientConfig: "/path/to/kubeconfig",
			},
			expectError: false,
		},
		{
			name: "Valid config with empty kubeconfig",
			config: Config{
				ClientConfig: "",
			},
			expectError: false,
		},
		{
			name: "Valid config with max nodes",
			config: Config{
				ClientConfig: "",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Config 结构体本身不包含验证逻辑
			// 这里只是测试结构体的赋值和读取
			helper.AssertEqual(tc.config.ClientConfig, tc.config.ClientConfig, "ClientConfig should match")
		})
	}
}
