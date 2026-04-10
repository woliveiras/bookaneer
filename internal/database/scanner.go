package database

// Scanner abstracts *sql.Row and *sql.Rows for shared row-scanning helpers.
// Defining it once here avoids duplicating the same one-method interface across packages.
type Scanner interface {
	Scan(dest ...any) error
}
