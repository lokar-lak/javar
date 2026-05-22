package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/glebarez/sqlite"
)

// Open opens SQLite DB and runs migration.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite works better with a single write connection
	db.SetMaxOpenConns(1)

	if err := ping(db); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func ping(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping sqlite: %w", err)
	}

	// Migration: add orthography column if it does not exist yet
	db.Exec(`ALTER TABLE translations ADD COLUMN orthography TEXT NOT NULL DEFAULT 'academic'`)
	return nil
}

func migrate(db *sql.DB) error {
	// Read SQL file and execute it
	data, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		// Ignore "already exists" errors on repeated runs
		// In real projects, use goose or migrate
		_ = err
	}

	// Backfill schema for existing databases: SteamDB source link.
	if _, err := db.Exec(`ALTER TABLE games ADD COLUMN steamdb_url TEXT NOT NULL DEFAULT ''`); err != nil {
		// SQLite returns duplicate column error if this column already exists.
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("add steamdb_url: %w", err)
		}
	}
	if _, err := db.Exec(`ALTER TABLE games ADD COLUMN platforms TEXT NOT NULL DEFAULT '[]'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("add games.platforms: %w", err)
		}
	}
	if _, err := db.Exec(`ALTER TABLE translations ADD COLUMN official_status TEXT NOT NULL DEFAULT 'unofficial'`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("add translations.official_status: %w", err)
		}
	}

	// Remove legacy page-specific cover column if present.
	if _, err := db.Exec(`ALTER TABLE games DROP COLUMN page_cover_url`); err != nil {
		errMsg := strings.ToLower(err.Error())
		if !strings.Contains(errMsg, "no such column") && !strings.Contains(errMsg, "syntax error") {
			return fmt.Errorf("drop page_cover_url: %w", err)
		}
	}

	// Remove legacy translations version column if present.
	if _, err := db.Exec(`ALTER TABLE translations DROP COLUMN version`); err != nil {
		errMsg := strings.ToLower(err.Error())
		if !strings.Contains(errMsg, "no such column") && !strings.Contains(errMsg, "syntax error") {
			return fmt.Errorf("drop translations.version: %w", err)
		}
	}

	// Remove legacy studio name from translations; authors are stored in translator_names.
	if _, err := db.Exec(`ALTER TABLE translations DROP COLUMN studio_name`); err != nil {
		errMsg := strings.ToLower(err.Error())
		if !strings.Contains(errMsg, "no such column") && !strings.Contains(errMsg, "syntax error") {
			return fmt.Errorf("drop translations.studio_name: %w", err)
		}
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS translation_submissions (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			game_title        TEXT    NOT NULL,
			platforms         TEXT    NOT NULL DEFAULT '[]',
			category          TEXT    NOT NULL CHECK(category IN ('official','unofficial')),
			localization_type TEXT    NOT NULL DEFAULT '[]',
			authors           TEXT    NOT NULL,
			game_url          TEXT    NOT NULL,
			translation_url   TEXT    NOT NULL,
			description       TEXT    NOT NULL DEFAULT '',
			status            TEXT    NOT NULL DEFAULT 'new' CHECK(status IN ('new','accepted','rejected')),
			created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("create translation_submissions: %w", err)
	}

	if err := seedSteamTagGenres(db); err != nil {
		return err
	}
	if err := normalizeTranslationCoverage(db); err != nil {
		return err
	}
	return nil
}

func normalizeTranslationCoverage(db *sql.DB) error {
	rows, err := db.Query(`SELECT id, coverage FROM translations`)
	if err != nil {
		return fmt.Errorf("list translation coverage: %w", err)
	}
	defer rows.Close()

	type item struct {
		id       int
		coverage string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.coverage); err != nil {
			return fmt.Errorf("scan translation coverage: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate translation coverage: %w", err)
	}

	for _, it := range items {
		var values []string
		json.Unmarshal([]byte(it.coverage), &values)
		normalized := normalizeCoverageValues(values)
		data, _ := json.Marshal(normalized)
		if string(data) == it.coverage {
			continue
		}
		if _, err := db.Exec(`UPDATE translations SET coverage=? WHERE id=?`, string(data), it.id); err != nil {
			return fmt.Errorf("update translation coverage %d: %w", it.id, err)
		}
	}
	return nil
}

func normalizeCoverageValues(values []string) []string {
	hasText := false
	hasVoice := false
	for _, value := range values {
		v := strings.ToLower(strings.TrimSpace(value))
		if v == "" {
			continue
		}
		if strings.Contains(v, "агуч") || strings.Contains(v, "озвуч") || strings.Contains(v, "voice") || strings.Contains(v, "audio") {
			hasVoice = true
			continue
		}
		hasText = true
	}

	var out []string
	if hasText {
		out = append(out, "Тэкст")
	}
	if hasVoice {
		out = append(out, "Агучка")
	}
	return out
}

func seedSteamTagGenres(db *sql.DB) error {
	genres := []struct {
		name string
		slug string
	}{
		{"Галаваломка", "puzzle"},
		{"Экшн-прыгода", "action-adventure"},
		{"Аркада", "arcade"},
		{"Шутар", "shooter"},
		{"Платформер", "platformer"},
		{"Візуальная навела", "visual-novel"},
		{"Роглайк", "roguelike"},
		{"Пясочніца", "sandbox"},
		{"Point-and-click", "point-and-click"},
		{"Экшн-RPG", "action-rpg"},
		{"Экшн-роглайк", "action-roguelike"},
		{"Інтэрактыўная літаратура", "interactive-fiction"},
		{"Пакрокавая стратэгія", "turn-based-strategy"},
		{"Настольныя", "tabletop"},
		{"Сімулятар хадзьбы", "walking-simulator"},
		{"Сімулятар спатканняў", "dating-sim"},
		{"Картачная гульня", "card-game"},
		{"JRPG", "jrpg"},
		{"Адукацыйная", "education"},
		{"Сімулятар жыцця", "life-sim"},
		{"Партыйная RPG", "party-based-rpg"},
		{"Стратэгічная RPG", "strategy-rpg"},
		{"Утыліты", "utilities"},
		{"Настольная гульня", "board-game"},
		{"Абарона вежаў", "tower-defense"},
		{"RTS", "rts"},
		{"Будаўніцтва горада", "city-builder"},
		{"Beat 'em up", "beat-em-up"},
		{"Аўтасімулятар", "automobile-sim"},
		{"2D-файтэр", "2d-fighter"},
		{"Сімулятар фермферства", "farming-sim"},
		{"Слоўная гульня", "word-game"},
		{"3D-файтэр", "3d-fighter"},
		{"Сімулятар калоніі", "colony-sim"},
		{"Гульня для вечарынак", "party-game"},
		{"Касмічны сімулятар", "space-sim"},
		{"Глабальная стратэгія", "grand-strategy"},
		{"Кіберспорт", "esports"},
		{"MMORPG", "mmorpg"},
		{"Каралеўская бітва", "battle-royale"},
		{"Гульня ў бога", "god-game"},
		{"Выжыванне ў адкрытым свеце", "open-world-survival-craft"},
		{"Відэавытворчасць", "video-production"},
		{"4X", "4x"},
		{"MOBA", "moba"},
		{"Сімулятар працы", "job-simulator"},
		{"Віктарына", "trivia"},
		{"Сацыяльная дэдукцыя", "social-deduction"},
		{"Сімулятар хобі", "hobby-sim"},
	}

	for _, g := range genres {
		if _, err := db.Exec(`INSERT OR IGNORE INTO genres (name, slug) VALUES (?, ?)`, g.name, g.slug); err != nil {
			return fmt.Errorf("seed genre %s: %w", g.slug, err)
		}
	}
	return nil
}
