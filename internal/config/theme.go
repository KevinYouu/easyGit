package config

import (
	"database/sql"
	"errors"
	"fmt"
)

// 主题模式合法值:与 internal/theme 的 Mode 常量字符串保持一致
// (theme_test.TestModesAlignWithConfig 防漂移)
const (
	ThemeAuto  = "auto"
	ThemeDark  = "dark"
	ThemeLight = "light"
)

// ErrInvalidTheme 主题值非法的哨兵错误(errors.Is 可判定)
var ErrInvalidTheme = errors.New("invalid theme")

var validThemes = map[string]bool{
	ThemeAuto:  true,
	ThemeDark:  true,
	ThemeLight: true,
}

// GetTheme 从数据库获取主题设置(auto/dark/light);未设置返回空字符串
// (调用方按自动检测处理)。
func GetTheme() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var theme string
	query := `SELECT value FROM settings WHERE key = 'theme'`
	err = db.QueryRow(query).Scan(&theme)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有设置,返回空字符串
			return "", nil
		}
		return "", err
	}

	return theme, nil
}

// SaveTheme 保存主题设置到数据库(合法值:auto/dark/light)。
// 非法值返回包装 ErrInvalidTheme 的错误;合法值 UPSERT 覆盖。
func SaveTheme(theme string) error {
	if !validThemes[theme] {
		return fmt.Errorf("%w: %q", ErrInvalidTheme, theme)
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// 使用 UPSERT 语法
	query := `INSERT INTO settings (key, value) VALUES ('theme', ?)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	_, err = db.Exec(query, theme)
	return err
}
