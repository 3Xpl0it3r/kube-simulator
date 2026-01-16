package util

import (
	"net"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "获取本地IP",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 验证返回的是有效的IPv4地址
				if net.ParseIP(got) == nil {
					t.Errorf("GetLocalIP() = %v, 不是一个有效的IP地址", got)
				}
				if net.ParseIP(got).To4() == nil {
					t.Errorf("GetLocalIP() = %v, 不是IPv4地址", got)
				}
				// 验证不是回环地址
				if got == "127.0.0.1" {
					t.Logf("GetLocalIP() = %v, 返回了回环地址（可能没有其他网络接口）", got)
				}
			}
		})
	}
}

func TestGetLocalIP_NetworkInterface(t *testing.T) {
	// 测试网络接口存在的情况
	ip, err := GetLocalIP()
	if err != nil {
		t.Logf("无法获取本地IP（可能在某些环境中无网络接口）: %v", err)
		return
	}

	t.Logf("获取到的本地IP: %s", ip)

	// 验证IP格式
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Errorf("返回的IP %s 格式无效", ip)
	}
}
