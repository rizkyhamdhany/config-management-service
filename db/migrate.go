package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func MigrateSQLFiles(db *sql.DB, dir string) error {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, n := range names {
		b, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("migrations %s: %w", n, err)
		}
	}
	return nil
}
