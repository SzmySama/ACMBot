package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateQQ(t *testing.T) {
	// 创建一个内存数据库用于测试
	_, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化全局数据库连接（假设你有一个设置 db 的函数）
	GetDBConnection()

	// 测试表迁移
	err = MigrateQQ()
	if err != nil {
		t.Errorf("MigrateQQ() failed: %v", err)
	}
}

func TestUnBindQQToCodeforcesName(t *testing.T) {
	// 准备测试数据
	user := QQUser{
		QQNumber:       1599949878,
		QQGroupNumber:  753586106,
		CodeforcesName: "keeping_running",
	}

	err := UnBindQQToCodeforcesName(user)
	if err != nil {
		t.Errorf("BindQQToCodeforcesName() failed: %v", err)
	}

}

func TestGetCodeforcesName(t *testing.T) {
	// 准备测试数据
	user := QQUser{
		QQNumber:       1599949878,
		QQGroupNumber:  753586106,
		CodeforcesName: "keeping_running",
	}

	// 删除测试数据
	err := BindQQToCodeforcesName(user)
	if err != nil {
		t.Fatalf("Failed to unbind QQ to Codeforces name: %v", err)
	}

	// 测试查询
	codeforcesName, err := GetCodeforcesName(1599949878)
	if err != nil {
		t.Errorf("GetCodefocesName() failed: %v", err)
	}

	// 校验结果
	if codeforcesName != "keeping_running" {
		t.Errorf("Expected Codeforces name 'keeping_running', got '%s'", codeforcesName)
	}
}

func TestReBindQQToCodeforcesName(t *testing.T) {
	// 准备测试数据
	user := QQUser{
		QQNumber:       1599949878,
		QQGroupNumber:  753586106,
		CodeforcesName: "keep_running",
	}

	err := ReBindQQToCodeforcesName(user)
	if err != nil {
		t.Errorf("ReBindQQToCodeforcesName() failed: %v", err)
	}

}
