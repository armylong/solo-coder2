package config

import (
	"testing"
)

func TestSyncEnv4System(t *testing.T) {
	// 测试写入环境变量到系统内环境变量
	err := SyncEnv4System("test_env", "test_value")
	if err != nil {
		t.Errorf("SyncEnv4System failed: %v", err)
	}
}
