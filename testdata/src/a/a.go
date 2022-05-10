package a

import (
	"context"
	"database/sql"
)

func sample(db *sql.DB) {
	s := "alice"

	_ = db.QueryRowContext(context.Background(), "SELECT * FROM test WHERE test=?", s)

	_ = db.QueryRowContext(context.Background(), "DELETE * FROM test WHERE test=?", s) // want "It\\'s better to use Execute method instead of QueryRowContext method to execute `DELETE` query"
	_ = db.QueryRowContext(context.Background(), "UPDATE * FROM test WHERE test=?", s) // want "It\\'s better to use Execute method instead of QueryRowContext method to execute `UPDATE` query"

	_, _ = db.Query("UPDATE * FROM test WHERE test=?", s)                              // want "It\\'s better to use Execute method instead of Query method to execute `UPDATE` query"
	_, _ = db.QueryContext(context.Background(), "UPDATE * FROM test WHERE test=?", s) // want "It\\'s better to use Execute method instead of QueryContext method to execute `UPDATE` query"
	_ = db.QueryRow("UPDATE * FROM test WHERE test=?", s)                              // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

	query := "UPDATE * FROM test where test=?"
	_, _ = db.Query(query, s) // want "It\\'s better to use Execute method instead of Query method to execute `UPDATE` query"

	var f1 = `
UPDATE * FROM test WHERE test=?`
	_ = db.QueryRow(f1, s) // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

	const f2 = `
UPDATE * FROM test WHERE test=?`
	_ = db.QueryRow(f2, s) // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

	f3 := `
UPDATE * FROM test WHERE test=?`
	_ = db.QueryRow(f3, s) // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

	f4 := f3
	_ = db.QueryRow(f4, s) // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

	f5 := `
UPDATE * ` + `FROM test` + ` WHERE test=?`
	_ = db.QueryRow(f5, s) // want "It\\'s better to use Execute method instead of QueryRow method to execute `UPDATE` query"

}
