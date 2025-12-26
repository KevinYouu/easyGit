package config

import (
	"fmt"
)

// get the default options
func GetDefaultOptions() []Option {
	return []Option{
		{Label: "fix", Value: "fix", Usage: 0},
		{Label: "feat", Value: "feat", Usage: 0},
		{Label: "refactor", Value: "refactor", Usage: 0},
		{Label: "build", Value: "build", Usage: 0},
		{Label: "chore", Value: "chore", Usage: 0},
		{Label: "style", Value: "style", Usage: 0},
		{Label: "docs", Value: "docs", Usage: 0},
		{Label: "revert", Value: "revert", Usage: 0},
		{Label: "test", Value: "test", Usage: 0},
	}
}

func GetOptions() ([]Option, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT label, value, usage FROM options ORDER BY usage DESC")
	if err != nil {
		if initErr := Initialize(); initErr != nil {
			return nil, fmt.Errorf("failed to initialize database: %v", initErr)
		}
		rows, err = db.Query("SELECT label, value, usage FROM options ORDER BY usage DESC")
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var options []Option
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.Label, &option.Value, &option.Usage); err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func SaveOptions(options []Option) error {
	return SaveRecords(options, "options", []string{"label", "value", "usage"}, "value", []string{"label", "usage"})
}
