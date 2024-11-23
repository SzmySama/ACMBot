package db

import (
	"fmt"
	"testing"
)

func TestBindQQAndCodeforcesHandler(t *testing.T) {
	// 初始化数据库

	// 自动迁移数据结构
	if err := db.AutoMigrate(&QQUser{}, &Group{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// 测试数据
	tests := []struct {
		qqNumber         uint
		groupNumber      uint
		codeforcesHandle string
		expectedError    string
	}{
		{123456, 1, "user1", ""},
		{123456, 1, "user1", ""},
		{123457, 1, "user1", "codeforces Handle user1 has binded by others"},
		{123456, 2, "user2", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("QQNumber=%v, GroupNumber=%v, Handle=%s", tt.qqNumber, tt.groupNumber, tt.codeforcesHandle), func(t *testing.T) {
			err := BindQQandCodeforcesHandler(tt.qqNumber, tt.groupNumber, tt.codeforcesHandle)
			if (err != nil && err.Error() != tt.expectedError) || (err == nil && tt.expectedError != "") {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}

	// 验证数据库中的数据是否符合预期
	var user QQUser
	if err := db.Where("qq_number = ?", 123456).First(&user).Error; err != nil {
		t.Errorf("failed to find QQUser: %v", err)
	}
	if user.CodeforcesHandle != "user2" {
		t.Errorf("expected codeforces handle 'user2', got: %v", user.CodeforcesHandle)
	}

	var group Group
	if err := db.Where("group_id = ?", 1).First(&group).Error; err != nil {
		t.Errorf("failed to find Group: %v", err)
	}
	if len(group.QQUsers) != 2 {
		t.Errorf("expected 2 QQUsers in group, got: %v", len(group.QQUsers))
	}
}
