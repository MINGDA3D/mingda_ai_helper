package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateMachineSN 验证机器序列号格式
func ValidateMachineSN(sn string) bool {
	// TODO: 实现机器序列号格式验证逻辑
	return len(sn) > 0
}

// ValidateMachineModel 验证机器型号格式
func ValidateMachineModel(model string) bool {
	// TODO: 实现机器型号格式验证逻辑
	return len(model) > 0
} 