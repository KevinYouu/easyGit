package config

import (
	"database/sql"
)

// ConflictEditorCustom 配置中心「自定义编辑器」选项值(选中后走输入子流程)
const ConflictEditorCustom = "__custom__"

// GetConflictEditor 从数据库获取冲突编辑器设置(编辑器命令,如 "nano");
// 未设置返回空字符串,调用方按自动检测($EDITOR → vim/vi/nano)处理。
func GetConflictEditor() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var editor string
	query := `SELECT value FROM settings WHERE key = 'conflict_editor'`
	err = db.QueryRow(query).Scan(&editor)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有设置,返回空字符串(自动检测)
			return "", nil
		}
		return "", err
	}

	return editor, nil
}

// SaveConflictEditor 保存冲突编辑器设置(任意编辑器命令字符串;
// 空串表示清除设置,恢复自动检测);UPSERT 覆盖。
func SaveConflictEditor(editor string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO settings (key, value) VALUES ('conflict_editor', ?)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	_, err = db.Exec(query, editor)
	return err
}
