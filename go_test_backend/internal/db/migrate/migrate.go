package migrate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migration struct {
	Version  string
	Filename string
	SQL      string
}

func ApplyDir(db *gorm.DB, dir string, log *zap.Logger) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	migs, err := loadMigrations(dir)
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		log.Info("no migrations found", zap.String("dir", dir))
		return nil
	}

	applied, err := getAppliedVersions(db)
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.Version] {
			continue
		}
		log.Info("applying migration", zap.String("version", m.Version), zap.String("file", m.Filename))
		if err := applyOne(db, m); err != nil {
			return fmt.Errorf("apply %s: %w", m.Filename, err)
		}
	}
	return nil
}

func ensureMigrationsTable(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version    TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL
);
`).Error
}

func getAppliedVersions(db *gorm.DB) (map[string]bool, error) {
	type row struct {
		Version string
	}
	var rows []row
	if err := db.Raw(`SELECT version FROM schema_migrations`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(rows))
	for _, r := range rows {
		out[r.Version] = true
	}
	return out, nil
}

func applyOne(db *gorm.DB, m Migration) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(m.SQL).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, m.Version, time.Now().UTC()).Error; err != nil {
			return err
		}
		return nil
	})
}

func loadMigrations(dir string) ([]Migration, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("migrations dir is not a directory")
	}

	var migs []Migration
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
			return nil
		}
		version, ok := parseVersion(d.Name())
		if !ok {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		migs = append(migs, Migration{
			Version:  version,
			Filename: d.Name(),
			SQL:      string(b),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].Version < migs[j].Version })
	return migs, nil
}

func parseVersion(filename string) (string, bool) {
	// Flyway-like: V1__desc.sql, V20260515_1__desc.sql, etc.
	base := filepath.Base(filename)
	if len(base) < 3 {
		return "", false
	}
	if base[0] != 'V' && base[0] != 'v' {
		return "", false
	}
	parts := strings.SplitN(base[1:], "__", 2)
	if len(parts) < 2 {
		return "", false
	}
	version := parts[0]
	if version == "" {
		return "", false
	}
	return version, true
}
