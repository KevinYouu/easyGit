package config

import (
	"database/sql"
)

// GetLanguage 从数据库获取语言设置
func GetLanguage() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var lang string
	query := `SELECT value FROM settings WHERE key = 'language'`
	err = db.QueryRow(query).Scan(&lang)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有设置,返回空字符串
			return "", nil
		}
		return "", err
	}

	return lang, nil
}

// SaveLanguage 保存语言设置到数据库
func SaveLanguage(lang string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// 使用 UPSERT 语法
	query := `INSERT INTO settings (key, value) VALUES ('language', ?)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	_, err = db.Exec(query, lang)
	return err
}
