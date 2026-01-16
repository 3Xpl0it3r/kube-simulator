package cluster

import (
	"testing"
)

func TestClusterConfig_DefaultValues(t *testing.T) {
	config := &Config{}

	if config.ListenHost != "" {
		t.Error("ListenHost should be empty by default")
	}

	if config.ListenPort != "" {
		t.Error("ListenPort should be empty by default")
	}
}

func TestClusterConfig_SetValues(t *testing.T) {
	config := &Config{
		ListenHost:        "127.0.0.1",
		ListenPort:        "8443",
		AuthorizationMode: "RBAC",
	}

	if config.ListenHost != "127.0.0.1" {
		t.Errorf("Expected ListenHost to be 127.0.0.1, got %s", config.ListenHost)
	}

	if config.ListenPort != "8443" {
		t.Errorf("Expected ListenPort to be 8443, got %s", config.ListenPort)
	}

	if config.AuthorizationMode != "RBAC" {
		t.Errorf("Expected AuthorizationMode to be RBAC, got %s", config.AuthorizationMode)
	}
}

func TestGetArgsList_BasicMap(t *testing.T) {
	argsMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	args := GetArgsList(argsMap, nil)

	// 验证返回的参数数量
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	// 验证参数格式和排序（应该按字母顺序排序）
	expectedArgs := []string{
		"--key1=value1",
		"--key2=value2",
		"--key3=value3",
	}

	for i, expected := range expectedArgs {
		if i >= len(args) || args[i] != expected {
			t.Errorf("Expected arg %d to be %s, got %s", i, expected, getArgAt(args, i))
		}
	}
}

func TestGetArgsList_WithExtraArgs(t *testing.T) {
	argsMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	extraArgs := []string{
		"key3=value3",
		"key4=value4",
		"key5=true",
	}

	args := GetArgsList(argsMap, extraArgs)

	// 验证返回的参数数量
	if len(args) != 5 {
		t.Errorf("Expected 5 args, got %d", len(args))
	}

	// 验证 extraArgs 覆盖了原有参数
	if getValueFromArgs(args, "key3") != "value3" {
		t.Errorf("Expected key3 to be value3, got %s", getValueFromArgs(args, "key3"))
	}
	if getValueFromArgs(args, "key5") != "true" {
		t.Errorf("Expected key5 to be true, got %s", getValueFromArgs(args, "key5"))
	}
}

func TestGetArgsList_EmptyMap(t *testing.T) {
	args := GetArgsList(map[string]string{}, nil)

	if len(args) != 0 {
		t.Errorf("Expected empty args list, got %d args", len(args))
	}
}

// 辅助函数：从参数列表中获取指定参数的值
func getValueFromArgs(args []string, key string) string {
	prefix := "--" + key + "="
	for _, arg := range args {
		if len(arg) > len(prefix) && arg[:len(prefix)] == prefix {
			return arg[len(prefix):]
		}
	}
	return ""
}

// 辅助函数：获取指定位置的参数
func getArgAt(args []string, index int) string {
	if index >= 0 && index < len(args) {
		return args[index]
	}
	return ""
}

func TestClusterLogging_Initialization(t *testing.T) {
	// 测试集群日志初始化
	// 验证日志记录器是否正确初始化

	// 这些是全局变量，我们只能验证它们不为 nil
	if loggerForApiServer == nil {
		t.Error("API Server logger should be initialized")
	}

	if loggerForScheduler == nil {
		t.Error("Scheduler logger should be initialized")
	}

	if loggerForControllerMg == nil {
		t.Error("Controller Manager logger should be initialized")
	}
}
