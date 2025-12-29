package config

import (
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// PushConfig 推送配置
type PushConfig struct {
	Remotes []string // 远程仓库名列表(支持多个)
}

// getRepoKey 获取当前仓库的唯一标识
func getRepoKey() (string, error) {
	// 获取当前仓库的根目录绝对路径作为唯一标识
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}

	repoPath := strings.TrimSpace(string(output))
	// 使用绝对路径确保唯一性
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// GetPushConfig 获取当前仓库的推送配置
func GetPushConfig() (*PushConfig, error) {
	repoKey, err := getRepoKey()
	if err != nil {
		return nil, err
	}

	db, err := openDB()
	if err != nil {
		return nil, err
	}

	// 使用仓库路径作为 key 前缀
	remotesKey := fmt.Sprintf("push_remotes:%s", repoKey)

	// 获取远程列表 (逗号分隔格式存储)
	var remotesStr string
	err = db.QueryRow("SELECT value FROM settings WHERE key = ?", remotesKey).Scan(&remotesStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get push_remotes: %w", err)
	}

	var remotes []string
	if remotesStr != "" {
		remotes = strings.Split(remotesStr, ",")
	}

	// 如果没有配置,返回 nil
	if len(remotes) == 0 {
		return nil, nil
	}

	return &PushConfig{
		Remotes: remotes,
	}, nil
}

// SavePushConfig 保存当前仓库的推送配置
func SavePushConfig(remotes []string) error {
	repoKey, err := getRepoKey()
	if err != nil {
		return err
	}

	db, err := openDB()
	if err != nil {
		return err
	}

	// 使用仓库路径作为 key 前缀
	remotesKey := fmt.Sprintf("push_remotes:%s", repoKey)

	// 保存远程列表 (逗号分隔格式)
	remotesStr := strings.Join(remotes, ",")

	_, err = db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, remotesKey, remotesStr)
	if err != nil {
		return fmt.Errorf("save push_remotes: %w", err)
	}

	return nil
}

// ClearPushConfig 清除当前仓库的推送配置
func ClearPushConfig() error {
	repoKey, err := getRepoKey()
	if err != nil {
		return err
	}

	db, err := openDB()
	if err != nil {
		return err
	}

	remotesKey := fmt.Sprintf("push_remotes:%s", repoKey)

	_, err = db.Exec("DELETE FROM settings WHERE key = ?", remotesKey)
	if err != nil {
		return fmt.Errorf("clear push config: %w", err)
	}

	return nil
}
