package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	_ "modernc.org/sqlite"
)

var testDBPath string // 测试用数据库路径

// SetTestDBPath 设置测试数据库路径(仅用于测试)
func SetTestDBPath(path string) {
	testDBPath = path
}

// Get the path to the SQLite database
func getDBPath() string {
	if testDBPath != "" {
		return testDBPath
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".easyGit.db")
}

// Open the SQLite database
func openDB() (*sql.DB, error) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SaveRecords[T any](
	records []T,
	tableName string,
	columns []string,
	conflictColumn string, // If conflictColumn is not empty, use ON CONFLICT
	updateColumns []string, // Columns that need to be updated during a conflict
) error {
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

	columnPlaceholders := make([]string, len(columns))
	for i := range columns {
		columnPlaceholders[i] = "?"
	}

	baseSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(columnPlaceholders, ", "),
	)

	if conflictColumn != "" {
		updateSet := make([]string, len(updateColumns))
		for i, col := range updateColumns {
			updateSet[i] = fmt.Sprintf("%s=excluded.%s", col, col)
		}
		baseSQL += fmt.Sprintf(" ON CONFLICT(%s) DO UPDATE SET %s", conflictColumn, strings.Join(updateSet, ", "))
	}

	stmt, err := tx.Prepare(baseSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		values := extractValues(record, columns)
		if _, err := stmt.Exec(values...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// extractValues is a helper function to extract values from a struct
func extractValues[T any](record T, columns []string) []any {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	caser := cases.Title(language.Und)

	values := make([]any, len(columns))
	for i, col := range columns {
		field := v.FieldByName(caser.String(col))
		if field.IsValid() {
			values[i] = field.Interface()
		} else {
			values[i] = nil
		}
	}
	return values
}

func Initialize() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create the options table
	optionsTable := `CREATE TABLE IF NOT EXISTS options (
		label TEXT,
		value TEXT PRIMARY KEY,
		usage INTEGER
	)`
	_, err = db.Exec(optionsTable)
	if err != nil {
		return err
	}

	// Create the patches table
	patchesTable := `CREATE TABLE IF NOT EXISTS patches (
		prefix TEXT,
		major INTEGER,
		minor INTEGER,
		patch INTEGER,
		suffix TEXT
	)`
	_, err = db.Exec(patchesTable)
	if err != nil {
		return err
	}

	// Create the settings table
	settingsTable := `CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	)`
	_, err = db.Exec(settingsTable)
	if err != nil {
		return err
	}

	// Count the number of options
	countQuery := `SELECT COUNT(*) FROM options`
	var count int
	err = db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		return err
	}

	// If no options exist, insert default options
	if count == 0 {
		defaultOptions := GetDefaultOptions()
		if err := SaveOptions(defaultOptions); err != nil {
			return err
		}
	}

	// Count the number of patches
	countQuery = `SELECT COUNT(*) FROM patches`
	err = db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		return err
	}

	// If no patches exist, insert default patches
	if count == 0 {
		defaultPatches := GetDefaultTagPatch()
		if err := SavePatches(defaultPatches); err != nil {
			return err
		}
	}

	return nil
}
