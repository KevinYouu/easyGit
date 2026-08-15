package config

import (
	"fmt"
	"regexp"
)

// commitTypePattern 提交类型合法格式:小写字母、数字、连字符
var commitTypePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// AddCommitType 新增提交类型(usage 初始 0)。
// 格式非法或已存在时返回错误;label 同时作为 label 与 value 存储。
func AddCommitType(label string) error {
	if !commitTypePattern.MatchString(label) {
		return fmt.Errorf("invalid commit type format: %q", label)
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM options WHERE value = ?", label).Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("commit type %q already exists", label)
	}

	_, err = db.Exec("INSERT INTO options (label, value, usage) VALUES (?, ?, 0)", label, label)
	if err != nil {
		return fmt.Errorf("add commit type: %w", err)
	}
	return nil
}

// DeleteCommitTypes 删除多个提交类型(按 value 精确匹配)。
// 删除后不允许清空全部类型(至少保留一个);不存在的 value 静默跳过。
func DeleteCommitTypes(values []string) error {
	if len(values) == 0 {
		return nil
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM options").Scan(&total); err != nil {
		return err
	}
	if total-len(values) < 1 {
		return fmt.Errorf("at least one commit type must remain")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM options WHERE value = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range values {
		if _, err := stmt.Exec(v); err != nil {
			return err
		}
	}

	return tx.Commit()
}
