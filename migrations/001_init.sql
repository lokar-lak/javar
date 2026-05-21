-- ============================================================
--  APÚKA — SQLite schema + example data
-- ============================================================

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- ── Genres ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS genres (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL,
    slug TEXT    NOT NULL UNIQUE
);

-- ── Games ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS games (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT    NOT NULL,
    slug         TEXT    NOT NULL UNIQUE,
    developer    TEXT    NOT NULL,
    publisher    TEXT,
    release_date TEXT,
    description  TEXT,
    cover_url    TEXT,
    steamdb_url  TEXT,
    platforms    TEXT    NOT NULL DEFAULT '[]',
    steam_rating INTEGER,                          -- percentage, e.g. 83
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ── Many-to-many: games ↔ genres ─────────────────────────────
CREATE TABLE IF NOT EXISTS game_genres (
    game_id  INTEGER NOT NULL REFERENCES games(id)  ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, genre_id)
);

-- ── Translations ──────────────────────────────────────────────
-- One game can have multiple translations from different studios.
-- translator_names and coverage are stored as JSON arrays in text fields.
CREATE TABLE IF NOT EXISTS translations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id          INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    studio_name      TEXT    NOT NULL,
    translator_names TEXT    NOT NULL DEFAULT '[]',  -- JSON: ["Ivan K.","Maria P."]
    type             TEXT    NOT NULL CHECK(type IN ('manual','ai')),
    official_status  TEXT    NOT NULL DEFAULT 'unofficial' CHECK(official_status IN ('official','unofficial')),
    coverage         TEXT    NOT NULL DEFAULT '[]',  -- JSON: ["subtitles","menu",...]
    external_url     TEXT    NOT NULL,
    click_count      INTEGER NOT NULL DEFAULT 0,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ── Reviews (anonymous) ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS reviews (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    translation_id INTEGER NOT NULL REFERENCES translations(id) ON DELETE CASCADE,
    author_name    TEXT    NOT NULL,
    rating         INTEGER NOT NULL CHECK(rating BETWEEN 1 AND 5),
    body           TEXT    NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ── User-submitted Belarusian localization proposals ─────────
CREATE TABLE IF NOT EXISTS translation_submissions (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    game_title       TEXT    NOT NULL,
    platforms        TEXT    NOT NULL DEFAULT '[]',
    category         TEXT    NOT NULL CHECK(category IN ('official','unofficial')),
    localization_type TEXT   NOT NULL DEFAULT '[]',
    authors          TEXT    NOT NULL,
    game_url         TEXT    NOT NULL,
    translation_url  TEXT    NOT NULL,
    description      TEXT    NOT NULL DEFAULT '',
    status           TEXT    NOT NULL DEFAULT 'new' CHECK(status IN ('new','accepted','rejected')),
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

