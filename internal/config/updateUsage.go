package config

func IncrementUsage(value string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Ensure transaction is rolled back on error

	// Prepare the statement to update the usage count
	stmt, err := tx.Prepare(`
		UPDATE options
		SET usage = usage + 1
		WHERE value = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(value)
	if err != nil {
		tx.Rollback() // Rollback on error
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
