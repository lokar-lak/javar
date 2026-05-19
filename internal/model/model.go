package model

import "time"

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Game struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Developer   string    `json:"developer"`
	Publisher   string    `json:"publisher,omitempty"`
	ReleaseDate string    `json:"release_date,omitempty"`
	Description string    `json:"description,omitempty"`
	CoverURL    string    `json:"cover_url,omitempty"`
	SteamDBURL  string    `json:"steamdb_url,omitempty"`
	SteamRating *int      `json:"steam_rating,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Genres      []Genre   `json:"genres,omitempty"`
	HasOnlyAI   bool      `json:"has_only_ai,omitempty"`
	BestRating  float64   `json:"best_rating"`
}

// GameDetail - game plus all its translations (for game page)
type GameDetail struct {
	Game
	Translations []TranslationDetail `json:"translations"`
}

type Translation struct {
	ID              int       `json:"id"`
	GameID          int       `json:"game_id"`
	StudioName      string    `json:"studio_name"`
	TranslatorNames []string  `json:"translator_names"`
	Type            string    `json:"type"` // "manual" | "ai"
	Coverage        []string  `json:"coverage"`
	ExternalURL     string    `json:"external_url"`
	Orthography     string    `json:"orthography"`  // "academic" | "tarashkevitsa" | "lacinka"
	ClickCount      int       `json:"click_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TranslationDetail - translation plus aggregated rating and reviews
type TranslationDetail struct {
	Translation
	AvgRating       *float64 `json:"avg_rating"`
	ReviewCount     int      `json:"review_count"`
	RatingBreakdown [5]int   `json:"rating_breakdown"` // index 0 = 1★, 4 = 5★
	Reviews         []Review `json:"reviews,omitempty"`
}

type Review struct {
	ID            int       `json:"id"`
	TranslationID int       `json:"translation_id"`
	AuthorName    string    `json:"author_name"`
	Rating        int       `json:"rating"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
}

// ── Requests ─────────────────────────────────────────────────

type CreateGameRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Developer   string `json:"developer"`
	Publisher   string `json:"publisher"`
	ReleaseDate string `json:"release_date"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	SteamDBURL  string `json:"steamdb_url"`
	SteamRating *int   `json:"steam_rating"`
	GenreIDs    []int  `json:"genre_ids"`
}

type CreateTranslationRequest struct {
	GameID          int      `json:"game_id"`
	StudioName      string   `json:"studio_name"`
	TranslatorNames []string `json:"translator_names"`
	Type            string   `json:"type"`
	Coverage        []string `json:"coverage"`
	ExternalURL     string   `json:"external_url"`
	Orthography     string   `json:"orthography"`
}

type CreateReviewRequest struct {
	TranslationID int    `json:"translation_id"`
	AuthorName    string `json:"author_name"`
	Rating        int    `json:"rating"`
	Body          string `json:"body"`
}

type GameFilter struct {
	Search  string
	GenreID int    // genre filter (ID)
	Type         string // "manual" | "ai" | ""
	Orthography  string // "academic" | "tarashkevitsa" | "lacinka" | ""
	Page         int
	Limit        int
}

// ── Admin ─────────────────────────────────────────────────────

type AdminStats struct {
	TotalGames         int          `json:"total_games"`
	TotalTranslations  int          `json:"total_translations"`
	TotalReviews       int          `json:"total_reviews"`
	TotalClicks        int          `json:"total_clicks"`
	ManualTranslations int          `json:"manual_translations"`
	AITranslations     int          `json:"ai_translations"`
	TopGames           []GameClicks `json:"top_games"`
}

type GameClicks struct {
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Clicks     int    `json:"clicks"`
	Reviews    int    `json:"reviews"`
}

type AdminReview struct {
	Review
	GameTitle   string `json:"game_title"`
	GameSlug    string `json:"game_slug"`
	StudioName  string `json:"studio_name"`
}

type AdminTranslation struct {
	TranslationDetail
	GameTitle string `json:"game_title"`
	GameSlug  string `json:"game_slug"`
}
