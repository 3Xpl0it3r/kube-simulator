package cluster

import (
	"context"
	"testing"
	"time"
)

// ClusterTestHelper 集群测试辅助器
type ClusterTestHelper struct {
	T *testing.T
}

// NewClusterTestHelper 创建集群测试辅助器
func NewClusterTestHelper(t *testing.T) *ClusterTestHelper {
	return &ClusterTestHelper{
		T: t,
	}
}

// CreateTestConfig 创建测试配置
func (h *ClusterTestHelper) CreateTestConfig() *Config {
	return &Config{
		ListenHost:            "127.0.0.1",
		ListenPort:            "8443",
		AuthorizationMode:     "RBAC",
		ServiceClusterIpRange: "10.96.0.0/12",
		EtcdServers:           "127.0.0.1:2379",
		ClusterCIDR:           "10.244.0.0/16",
		ServiceCIDR:           "10.96.0.0/12",
		ClientConfigFile: ClientConfigFile{
			ControllerManager: "/tmp/controller-manager.kubeconfig",
			Scheduler:         "/tmp/scheduler.kubeconfig",
			Administrator:     "/tmp/admin.kubeconfig",
		},
	}
}

// AssertNoError 断言没有错误
func (h *ClusterTestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.T.Fatalf("%s: %v", msg, err)
	}
}

// AssertError 断言有错误
func (h *ClusterTestHelper) AssertError(err error, msg string) {
	if err == nil {
		h.T.Fatalf("Expected error but got none: %s", msg)
	}
}

// AssertEqual 断言相等
func (h *ClusterTestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.T.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// WaitForCondition 等待条件满足
func (h *ClusterTestHelper) WaitForCondition(condition func() bool, timeout time.Duration, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.T.Fatalf("Timeout waiting for condition: %s", msg)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
