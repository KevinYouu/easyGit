package config

import (
	"time"
)

// recentMessagesTable 最近提交消息表:独立表存储集合型数据
// (与 options 表同风格),数据库层保证去重(UNIQUE)、排序与截断,
// 避免 JSON 序列化塞进 settings 键值表。
const recentMessagesTable = `
CREATE TABLE IF NOT EXISTS recent_messages (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	message    TEXT NOT NULL UNIQUE,
	created_at INTEGER NOT NULL
)`

// RecentMessagesLimit 记忆的提交消息条数(↑/↓ 导航仍保持轻量)
const RecentMessagesLimit = 10

// GetRecentCommitMessages 读取最近提交消息(最新在前,最多 RecentMessagesLimit 条)。
// 表不存在或读取失败时返回空列表(优雅降级,不影响主流程)。
func GetRecentCommitMessages() []string {
	db, err := openDB()
	if err != nil {
		return nil
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT message FROM recent_messages ORDER BY created_at DESC, id DESC LIMIT ?",
		RecentMessagesLimit,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			return nil
		}
		messages = append(messages, message)
	}
	return messages
}

// AddRecentCommitMessage 记录一条提交消息:重复消息删除后重插
// (获得新 id 与时间戳,必然置顶);超出上限删除最旧记录。
// 事务包裹保证删除/重插/截断原子性。
func AddRecentCommitMessage(message string) error {
	if message == "" {
		return nil
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// 去重置顶:删除旧记录再插入,新 id + 新时间戳必然排在最前
	if _, err := tx.Exec("DELETE FROM recent_messages WHERE message = ?", message); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO recent_messages (message, created_at) VALUES (?, ?)", message, now); err != nil {
		return err
	}

	// 截断:只保留最近 RecentMessagesLimit 条
	if _, err := tx.Exec(
		"DELETE FROM recent_messages WHERE id NOT IN ("+
			"SELECT id FROM recent_messages ORDER BY created_at DESC, id DESC LIMIT ?)",
		RecentMessagesLimit,
	); err != nil {
		return err
	}

	return tx.Commit()
}
