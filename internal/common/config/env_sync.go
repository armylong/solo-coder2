package config

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/armylong/go-library/service/redis"
)

// Redis哈希key，存储所有需要同步到环境变量的配置
const EnvSyncRedisKey = "env_sync_config"

// 环境变量配置项
// value为配置值，encrypt为加密方式（空字符串表示明文，"aes"表示AES加密）
type EnvConfigItem struct {
	Value   string `json:"value"`   // 配置值（加密时为密文）
	Encrypt string `json:"encrypt"` // 加密方式: "" 明文, "aes" AES加密
}

var envAesKey = []byte("env_sync_key_16b")

// 从Redis同步环境变量
// 读取env_sync_config哈希，根据encrypt字段解密后设置到环境变量
func SyncEnvFromRedis() {
	redisClient := redis.GetClient("default")
	ctx := context.Background()

	result, err := redisClient.HGetAll(ctx, EnvSyncRedisKey).Result()
	if err != nil {
		fmt.Printf("同步环境变量失败: %v\n", err)
		return
	}

	for envKey, jsonStr := range result {
		var item EnvConfigItem
		if err := json.Unmarshal([]byte(jsonStr), &item); err != nil {
			fmt.Printf("解析环境变量配置失败[%s]: %v\n", envKey, err)
			continue
		}

		value := item.Value
		if item.Encrypt == "aes" && value != "" {
			decrypted, err := envAesDecrypt(value)
			if err != nil {
				fmt.Printf("解密环境变量失败[%s]: %v\n", envKey, err)
				continue
			}
			value = decrypted
		}

		os.Setenv(envKey, value)
	}
}

// 写入环境变量到系统内环境变量
func SyncEnv4System(envKey, value string) error {
	// 获取用户使用的shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		return errors.New("SHELL环境变量未设置")
	}
	fmt.Printf("shell: %s\n", shell)

	// 获取用户家目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户家目录失败: %w", err)
	}
	fmt.Printf("homeDir: %s\n", homeDir)

	shellRcMap := map[string][]string{
		"/bin/bash": {".bashrc", ".bash_profile", ".profile"},
		"/bin/zsh":  {".zshrc", ".zprofile"},
		"/bin/sh":   {".profile"},
	}
	shellRcFileNames := shellRcMap[shell]
	if len(shellRcFileNames) == 0 {
		return fmt.Errorf("不支持的shell: %s", shell)
	}

	desiredLine := fmt.Sprintf("export %s=%q", envKey, value)
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*(export\s+)?%s=.*$`, regexp.QuoteMeta(envKey)))
	firstExistingPath := ""

	for _, configFileName := range shellRcFileNames {
		shellRcPath := filepath.Join(homeDir, configFileName)
		raw, readErr := os.ReadFile(shellRcPath)
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("读取 shell 配置文件失败: %w", readErr)
		}

		if firstExistingPath == "" {
			firstExistingPath = shellRcPath
		}

		content := string(raw)
		matchedLine := pattern.FindString(content)
		if matchedLine == "" {
			continue
		}
		if strings.TrimSpace(matchedLine) == desiredLine {
			return nil
		}

		newContent := pattern.ReplaceAllString(content, desiredLine)
		if err := os.WriteFile(shellRcPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("覆盖 shell 配置文件中的环境变量失败: %w", err)
		}
		return nil
	}

	targetPath := firstExistingPath
	if targetPath == "" {
		targetPath = filepath.Join(homeDir, shellRcFileNames[0])
	}

	raw, readErr := os.ReadFile(targetPath)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("读取 shell 配置文件失败: %w", readErr)
	}

	f, openErr := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr != nil {
		return fmt.Errorf("打开 shell 配置文件失败: %w", openErr)
	}
	defer f.Close()

	prefix := ""
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		prefix = "\n"
	}
	if len(raw) == 0 {
		prefix = ""
	}
	if _, err := f.WriteString(prefix + desiredLine + "\n"); err != nil {
		return fmt.Errorf("写入环境变量到 shell 配置文件失败: %w", err)
	}

	return nil
}

// 保存配置到Redis
// 自动处理加密，写入env_sync_config哈希
func SaveEnvConfig(envKey, value, encrypt string) error {
	redisClient := redis.GetClient("default")
	ctx := context.Background()

	item := EnvConfigItem{Encrypt: encrypt}

	if encrypt == "aes" && value != "" {
		encrypted, err := envAesEncrypt(value)
		if err != nil {
			return fmt.Errorf("加密失败: %w", err)
		}
		item.Value = encrypted
	} else {
		item.Value = value
	}

	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	return redisClient.HSet(ctx, EnvSyncRedisKey, envKey, string(jsonBytes)).Err()
}

// 删除配置
func DeleteEnvConfig(envKeys ...string) error {
	redisClient := redis.GetClient("default")
	ctx := context.Background()
	return redisClient.HDel(ctx, EnvSyncRedisKey, envKeys...).Err()
}

// 获取配置（解密后的值）
func GetEnvConfig(envKey string) (string, error) {
	redisClient := redis.GetClient("default")
	ctx := context.Background()

	jsonStr, err := redisClient.HGet(ctx, EnvSyncRedisKey, envKey).Result()
	if err != nil {
		return "", err
	}

	var item EnvConfigItem
	if err := json.Unmarshal([]byte(jsonStr), &item); err != nil {
		return "", err
	}

	if item.Encrypt == "aes" && item.Value != "" {
		return envAesDecrypt(item.Value)
	}

	return item.Value, nil
}

// 获取配置（掩码后的值，用于展示）
func GetEnvConfigMasked(envKey string) string {
	value, err := GetEnvConfig(envKey)
	if err != nil || value == "" {
		return ""
	}
	return maskValue(value)
}

func maskValue(value string) string {
	if len(value) <= 7 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-3:]
}

func envAesEncrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(envAesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func envAesDecrypt(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(envAesKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文太短")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
