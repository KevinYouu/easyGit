package config

import (
	"database/sql"
	"errors"
	"fmt"
)

// 推送前 pull 策略取值
const (
	PullBeforePushAlways = "always" // 推送前先 pull(默认)
	PullBeforePushNever  = "never"  // 跳过 pull
)

// ErrInvalidPullSetting 非法 pull 策略值
var ErrInvalidPullSetting = errors.New("invalid pull-before-push setting")

// GetPullBeforePush 读取推送前 pull 策略(未设置默认 always)
func GetPullBeforePush() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM settings WHERE key = 'pull_before_push'").Scan(&value)
	if err == sql.ErrNoRows {
		return PullBeforePushAlways, nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SavePullBeforePush 保存推送前 pull 策略(仅接受 always/never)
func SavePullBeforePush(v string) error {
	if v != PullBeforePushAlways && v != PullBeforePushNever {
		return fmt.Errorf("%w: %s", ErrInvalidPullSetting, v)
	}
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO settings (key, value) VALUES ('pull_before_push', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		v,
	)
	return err
}
