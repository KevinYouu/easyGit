package config

import (
	"errors"
	"fmt"
	"regexp"
)

// CommitTypePattern 提交类型合法格式:小写字母、数字、连字符。
// 导出供命令层预校验复用(单一事实来源,规则变更仅需同步此处)。
var CommitTypePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// ErrCommitTypeExists 提交类型已存在的哨兵错误(errors.Is 可判定)
var ErrCommitTypeExists = errors.New("commit type already exists")

// AddCommitType 新增提交类型(usage 初始 0)。
// 格式非法返回普通错误,已存在返回包装 ErrCommitTypeExists 的错误;
// label 同时作为 label 与 value 存储。
func AddCommitType(label string) error {
	if !CommitTypePattern.MatchString(label) {
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
		return fmt.Errorf("%w: %q", ErrCommitTypeExists, label)
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
