package config

import (
	"database/sql"
	"errors"
	"fmt"
)

// 补录提交正文开关取值
const (
	CommitBodyOn  = "on"  // 主题确认后可选填多行正文
	CommitBodyOff = "off" // 只录单行主题(默认)
)

// ErrInvalidCommitBodySetting 非法补录正文开关值
var ErrInvalidCommitBodySetting = errors.New("invalid commit-body setting")

// GetCommitBody 读取补录提交正文开关(未设置默认关闭)
func GetCommitBody() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM settings WHERE key = 'commit_body'").Scan(&value)
	if err == sql.ErrNoRows {
		return CommitBodyOff, nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SaveCommitBody 保存补录提交正文开关(仅接受 on/off)
func SaveCommitBody(v string) error {
	if v != CommitBodyOn && v != CommitBodyOff {
		return fmt.Errorf("%w: %s", ErrInvalidCommitBodySetting, v)
	}
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO settings (key, value) VALUES ('commit_body', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		v,
	)
	return err
}
