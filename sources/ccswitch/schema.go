package ccswitch

import "fmt"

const expectedUserVersion = 10

// RequiredTables lists the minimum tables APIHub needs from cc-switch.db.
var RequiredTables = []string{
	"providers",
	"proxy_request_logs",
	"model_pricing",
	"provider_endpoints",
}

// ValidateSchema checks PRAGMA user_version and that all required tables exist.
// Returns a descriptive error if the cc-switch version is incompatible.
func (r *Reader) ValidateSchema() error {
	var version int
	if err := r.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("ccswitch: read user_version: %w", err)
	}
	if version < expectedUserVersion {
		return fmt.Errorf("ccswitch: unsupported schema version %d (need >= %d, upgrade cc-switch)",
			version, expectedUserVersion)
	}

	for _, tbl := range RequiredTables {
		var n int
		if err := r.db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			tbl,
		).Scan(&n); err != nil {
			return fmt.Errorf("ccswitch: check table %q: %w", tbl, err)
		}
		if n == 0 {
			return fmt.Errorf("ccswitch: required table %q is missing", tbl)
		}
	}

	return nil
}

// TableColumns returns the column names for a given table.
func (r *Reader) TableColumns(table string) ([]string, error) {
	rows, err := r.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var cid int
		var name, colType string
		var notnull int
		var dfltValue sqlNullStr
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

type sqlNullStr struct {
	s string
}

func (ns *sqlNullStr) Scan(value any) error {
	if value == nil {
		ns.s = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		ns.s = v
	case []byte:
		ns.s = string(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}
