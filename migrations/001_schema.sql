CREATE TABLE IF NOT EXISTS genres (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS games (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT NOT NULL,
    slug         TEXT NOT NULL UNIQUE,
    developer    TEXT NOT NULL,
    publisher    TEXT NOT NULL DEFAULT '',
    release_date TEXT NOT NULL DEFAULT '',
    description  TEXT NOT NULL DEFAULT '',
    cover_url    TEXT NOT NULL DEFAULT '',
    steamdb_url  TEXT NOT NULL DEFAULT '',
    platforms    TEXT NOT NULL DEFAULT '[]',
    steam_rating INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS game_genres (
    game_id  INTEGER NOT NULL REFERENCES games(id)  ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (game_id, genre_id)
);

CREATE TABLE IF NOT EXISTS translations (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id          INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    game_title       TEXT    NOT NULL DEFAULT '',
    studio           TEXT    NOT NULL DEFAULT '',
    translator_names TEXT    NOT NULL DEFAULT '[]',  -- JSON array
    type             TEXT    NOT NULL CHECK(type IN ('manual','ai')),
    official_status  TEXT    NOT NULL DEFAULT 'unofficial' CHECK(official_status IN ('official','semi-official','unofficial')),
    coverage         TEXT    NOT NULL DEFAULT '[]',  -- JSON array
    external_url     TEXT    NOT NULL,
    click_count      INTEGER NOT NULL DEFAULT 0,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS translations_game_title_insert
AFTER INSERT ON translations
BEGIN
    UPDATE translations
    SET game_title = (SELECT title FROM games WHERE id = NEW.game_id)
    WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS translations_game_title_game_update
AFTER UPDATE OF game_id ON translations
BEGIN
    UPDATE translations
    SET game_title = (SELECT title FROM games WHERE id = NEW.game_id)
    WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS games_title_update_translations
AFTER UPDATE OF title ON games
BEGIN
    UPDATE translations
    SET game_title = NEW.title
    WHERE game_id = NEW.id;
END;

CREATE TABLE IF NOT EXISTS reviews (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    translation_id INTEGER NOT NULL REFERENCES translations(id) ON DELETE CASCADE,
    reviewer_id    TEXT    NOT NULL DEFAULT '',
    author_name    TEXT    NOT NULL DEFAULT 'Ананім',
    rating         INTEGER NOT NULL CHECK(rating BETWEEN 1 AND 5),
    body           TEXT    NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reviews_one_per_reviewer_translation
ON reviews(translation_id, reviewer_id)
WHERE reviewer_id <> '';

CREATE TABLE IF NOT EXISTS review_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    reviewer_id TEXT NOT NULL,
    ip          TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_review_events_rate_limit
ON review_events(reviewer_id, ip, created_at);

-- Auto-update updated_at in translations
CREATE TRIGGER IF NOT EXISTS translations_updated_at
AFTER UPDATE ON translations
BEGIN
    UPDATE translations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Migration: translation orthography
ALTER TABLE translations ADD COLUMN IF NOT EXISTS
  orthography TEXT NOT NULL DEFAULT 'academic'
  CHECK(orthography IN ('academic','tarashkevitsa','lacinka'));
