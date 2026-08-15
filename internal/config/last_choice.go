package config

import (
	"database/sql"
)

// 上次选择记忆键:reset 模式 / merge 策略 / cherry-pick 选项。
// 各命令执行成功后保存,下次进入对应选择列表时预选,
// 减少高频重复决策(默认项通常就够)。
const (
	LastChoiceResetMode        = "last_choice_reset_mode"
	LastChoiceMergeStrategy    = "last_choice_merge_strategy"
	LastChoiceCherryPickOption = "last_choice_cherry_pick_option"
)

// GetLastChoice 读取上次选择(未设置时返回空串)
func GetLastChoice(key string) (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SaveLastChoice 保存上次选择(空值不保存,避免覆盖已有记忆)
func SaveLastChoice(key, value string) error {
	if value == "" {
		return nil
	}
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}
