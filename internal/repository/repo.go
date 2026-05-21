package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"javar/internal/model"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db: db} }

// ═══ GENRES ══════════════════════════════════════════════════════════════

func (r *Repo) ListGenres() ([]model.Genre, error) {
	rows, err := r.db.Query(`SELECT id, name, slug FROM genres ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ═══ GAMES ═══════════════════════════════════════════════════════════════

func (r *Repo) ListGames(f model.GameFilter) ([]model.Game, error) {
	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}

	q := `SELECT DISTINCT g.id, g.title, g.slug, g.developer,
		       COALESCE(g.publisher,''), COALESCE(g.release_date,''),
		       COALESCE(g.description,''), COALESCE(g.cover_url,''), COALESCE(g.steamdb_url,''),
		       g.steam_rating, g.created_at,
		       COALESCE(
		         (SELECT MAX(avg_r) FROM (
		           SELECT AVG(rv.rating) as avg_r
		           FROM translations t2
		           LEFT JOIN reviews rv ON rv.translation_id = t2.id
		           WHERE t2.game_id = g.id
		           GROUP BY t2.id
		         )),
		         0
		       ) as best_rating
		FROM games g
		LEFT JOIN game_genres gg ON gg.game_id = g.id
		LEFT JOIN translations t  ON t.game_id  = g.id
		WHERE 1=1`

	var args []any

	if f.Search != "" {
		q += ` AND (g.title LIKE ? OR g.developer LIKE ?)`
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}
	if f.GenreID > 0 {
		q += ` AND gg.genre_id = ?`
		args = append(args, f.GenreID)
	}
	if f.Type != "" {
		q += ` AND t.type = ?`
		args = append(args, f.Type)
	}
	if f.Orthography != "" {
		q += ` AND t.orthography = ?`
		args = append(args, f.Orthography)
	}

	// ── Sorting ──────────────────────────────────────────────────────────
	sortBy := "g.created_at"
	sortOrder := "DESC"
	switch f.SortBy {
	case "steam_rating":
		sortBy = "g.steam_rating"
	case "best_rating":
		sortBy = "best_rating"
	}
	if f.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	q += ` ORDER BY ` + sortBy + ` ` + sortOrder + ` NULLS LAST, g.id DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Page*f.Limit)

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var rating sql.NullInt64
		if err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Developer, &g.Publisher,
			&g.ReleaseDate, &g.Description, &g.CoverURL, &g.SteamDBURL, &rating, &g.CreatedAt, &g.BestRating); err != nil {
			return nil, err
		}
		if rating.Valid {
			v := int(rating.Int64)
			g.SteamRating = &v
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range games {
		games[i].Genres, _ = r.genresByGame(games[i].ID)
		games[i].HasOnlyAI = r.hasOnlyAI(games[i].ID)
	}
	return games, nil
}

func (r *Repo) GetGameBySlug(slug string) (*model.GameDetail, error) {
	var g model.Game
	var rating sql.NullInt64
	err := r.db.QueryRow(`
		SELECT id, title, slug, developer,
		       COALESCE(publisher,''), COALESCE(release_date,''),
		       COALESCE(description,''), COALESCE(cover_url,''),
		       COALESCE(steamdb_url,''), steam_rating, created_at
		FROM games WHERE slug = ?`, slug).Scan(
		&g.ID, &g.Title, &g.Slug, &g.Developer, &g.Publisher,
		&g.ReleaseDate, &g.Description, &g.CoverURL, &g.SteamDBURL,
		&rating, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if rating.Valid {
		v := int(rating.Int64)
		g.SteamRating = &v
	}
	g.Genres, _ = r.genresByGame(g.ID)
	g.HasOnlyAI = r.hasOnlyAI(g.ID)

	translations, err := r.translationsByGame(g.ID)
	if err != nil {
		return nil, err
	}
	return &model.GameDetail{Game: g, Translations: translations}, nil
}

func (r *Repo) CreateGame(req model.CreateGameRequest) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO games (title,slug,developer,publisher,release_date,description,cover_url,steamdb_url,steam_rating)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		req.Title, req.Slug, req.Developer, req.Publisher,
		req.ReleaseDate, req.Description, req.CoverURL, req.SteamDBURL, req.SteamRating)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	for _, gid := range req.GenreIDs {
		r.db.Exec(`INSERT INTO game_genres (game_id,genre_id) VALUES (?,?)`, id, gid)
	}
	return id, nil
}

func (r *Repo) UpdateGame(id int, req model.CreateGameRequest) error {
	_, err := r.db.Exec(`
		UPDATE games SET title=?,slug=?,developer=?,publisher=?,
		  release_date=?,description=?,cover_url=?,steamdb_url=?,steam_rating=?
		WHERE id=?`,
		req.Title, req.Slug, req.Developer, req.Publisher,
		req.ReleaseDate, req.Description, req.CoverURL, req.SteamDBURL, req.SteamRating, id)
	if err != nil {
		return err
	}
	r.db.Exec(`DELETE FROM game_genres WHERE game_id=?`, id)
	for _, gid := range req.GenreIDs {
		r.db.Exec(`INSERT INTO game_genres (game_id,genre_id) VALUES (?,?)`, id, gid)
	}
	return nil
}

func (r *Repo) DeleteGame(id int) error {
	_, err := r.db.Exec(`DELETE FROM games WHERE id=?`, id)
	return err
}

func (r *Repo) hasOnlyAI(gameID int) bool {
	var total, manual int
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=?`, gameID).Scan(&total)
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=? AND type='manual'`, gameID).Scan(&manual)
	return total > 0 && manual == 0
}

func (r *Repo) bestRating(gameID int) float64 {
	var rating sql.NullFloat64
	r.db.QueryRow(`
		SELECT COALESCE(MAX(avg_rating), 0)
		FROM (
			SELECT t.id,
			       CASE WHEN COUNT(rv.id) > 0 THEN AVG(rv.rating) ELSE 0 END as avg_rating
			FROM translations t
			LEFT JOIN reviews rv ON rv.translation_id = t.id
			WHERE t.game_id = ?
			GROUP BY t.id
		)`, gameID).Scan(&rating)
	if rating.Valid {
		return rating.Float64
	}
	return 0
}

func (r *Repo) genresByGame(gameID int) ([]model.Genre, error) {
	rows, err := r.db.Query(`
		SELECT gr.id, gr.name, gr.slug FROM genres gr
		JOIN game_genres gg ON gg.genre_id=gr.id
		WHERE gg.game_id=?`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ═══ TRANSLATIONS ════════════════════════════════════════════════════════

func (r *Repo) translationsByGame(gameID int) ([]model.TranslationDetail, error) {
	rows, err := r.db.Query(`
		SELECT id, game_id, studio_name, translator_names, type,
		       COALESCE(orthography,'academic'),
		       coverage, external_url, click_count, created_at, updated_at
		FROM translations WHERE game_id=? ORDER BY created_at DESC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.TranslationDetail
	for rows.Next() {
		var t model.Translation
		var namesJSON, coverageJSON string
		if err := rows.Scan(&t.ID, &t.GameID, &t.StudioName, &namesJSON, &t.Type,
			&t.Orthography, &coverageJSON, &t.ExternalURL,
			&t.ClickCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(namesJSON), &t.TranslatorNames)
		json.Unmarshal([]byte(coverageJSON), &t.Coverage)

		out = append(out, model.TranslationDetail{Translation: t})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		tid := out[i].ID

		var avg float64
		r.db.QueryRow(
			`SELECT COUNT(*), COALESCE(AVG(rating),0) FROM reviews WHERE translation_id=?`, tid,
		).Scan(&out[i].ReviewCount, &avg)
		if out[i].ReviewCount > 0 {
			out[i].AvgRating = &avg
		}

		for star := 1; star <= 5; star++ {
			var cnt int
			r.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE translation_id=? AND rating=?`, tid, star).Scan(&cnt)
			out[i].RatingBreakdown[star-1] = cnt
		}

		out[i].Reviews, _ = r.ReviewsByTranslation(tid)
	}
	return out, nil
}

func (r *Repo) CreateTranslation(req model.CreateTranslationRequest) (int64, error) {
	if req.Orthography == "" {
		req.Orthography = "academic"
	}
	namesJSON, _ := json.Marshal(req.TranslatorNames)
	coverageJSON, _ := json.Marshal(req.Coverage)
	res, err := r.db.Exec(`
		INSERT INTO translations (game_id,studio_name,translator_names,type,orthography,coverage,external_url)
		VALUES (?,?,?,?,?,?,?)`,
		req.GameID, req.StudioName, string(namesJSON),
		req.Type, req.Orthography, string(coverageJSON), req.ExternalURL)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) UpdateTranslation(id int, req model.CreateTranslationRequest) error {
	if req.Orthography == "" {
		req.Orthography = "academic"
	}
	namesJSON, _ := json.Marshal(req.TranslatorNames)
	coverageJSON, _ := json.Marshal(req.Coverage)
	_, err := r.db.Exec(`
		UPDATE translations SET studio_name=?,translator_names=?,type=?,orthography=?,
		  coverage=?,external_url=?,updated_at=? WHERE id=?`,
		req.StudioName, string(namesJSON), req.Type, req.Orthography,
		string(coverageJSON), req.ExternalURL, time.Now(), id)
	return err
}

func (r *Repo) DeleteTranslation(id int) error {
	_, err := r.db.Exec(`DELETE FROM translations WHERE id=?`, id)
	return err
}

func (r *Repo) IncrementClick(id int) (string, error) {
	var url string
	if err := r.db.QueryRow(`SELECT external_url FROM translations WHERE id=?`, id).Scan(&url); err != nil {
		return "", fmt.Errorf("not found")
	}
	r.db.Exec(`UPDATE translations SET click_count=click_count+1 WHERE id=?`, id)
	return url, nil
}

// ═══ REVIEWS ═════════════════════════════════════════════════════════════

func (r *Repo) ReviewsByTranslation(id int) ([]model.Review, error) {
	rows, err := r.db.Query(`
		SELECT id, translation_id, author_name, rating, body, created_at
		FROM reviews WHERE translation_id=? ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.TranslationID, &rv.AuthorName, &rv.Rating, &rv.Body, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) CreateReview(req model.CreateReviewRequest) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO reviews (translation_id,author_name,rating,body) VALUES (?,?,?,?)`,
		req.TranslationID, strings.TrimSpace(req.AuthorName), req.Rating, strings.TrimSpace(req.Body))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) FindSimilarGames(title string) ([]model.SimilarGame, error) {
	title = strings.TrimSpace(title)
	if len([]rune(title)) < 3 {
		return nil, nil
	}
	rows, err := r.db.Query(`
		SELECT title, slug FROM games
		WHERE LOWER(title) LIKE LOWER(?) OR LOWER(?) LIKE '%' || LOWER(title) || '%'
		ORDER BY title LIMIT 5`, "%"+title+"%", title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.SimilarGame
	for rows.Next() {
		var g model.SimilarGame
		if err := rows.Scan(&g.Title, &g.Slug); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) CreateTranslationSubmission(req model.CreateTranslationSubmissionRequest) (int64, error) {
	platformsJSON, _ := json.Marshal(req.Platforms)
	typesJSON, _ := json.Marshal(req.LocalizationType)
	res, err := r.db.Exec(`
		INSERT INTO translation_submissions
		(game_title, platforms, category, localization_type, authors, game_url, translation_url, description)
		VALUES (?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(req.GameTitle), string(platformsJSON), req.Category, string(typesJSON),
		strings.TrimSpace(req.Authors), strings.TrimSpace(req.GameURL),
		strings.TrimSpace(req.TranslationURL), strings.TrimSpace(req.Description))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) DeleteReview(id int) error {
	_, err := r.db.Exec(`DELETE FROM reviews WHERE id=?`, id)
	return err
}

// ═══ ADMIN ═══════════════════════════════════════════════════════════════

func (r *Repo) GetStats() (*model.AdminStats, error) {
	s := &model.AdminStats{}
	r.db.QueryRow(`SELECT COUNT(*) FROM games`).Scan(&s.TotalGames)
	r.db.QueryRow(`SELECT COUNT(*) FROM translations`).Scan(&s.TotalTranslations)
	r.db.QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&s.TotalReviews)
	r.db.QueryRow(`SELECT COALESCE(SUM(click_count),0) FROM translations`).Scan(&s.TotalClicks)
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE type='manual'`).Scan(&s.ManualTranslations)
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE type='ai'`).Scan(&s.AITranslations)

	rows, err := r.db.Query(`
		SELECT g.title, g.slug,
		       COALESCE(SUM(t.click_count),0) as clicks,
		       COUNT(DISTINCT rv.id) as reviews
		FROM games g
		LEFT JOIN translations t ON t.game_id = g.id
		LEFT JOIN reviews rv ON rv.translation_id = t.id
		GROUP BY g.id
		ORDER BY clicks DESC
		LIMIT 10`)
	if err != nil {
		return s, nil
	}
	defer rows.Close()
	for rows.Next() {
		var gc model.GameClicks
		if err := rows.Scan(&gc.Title, &gc.Slug, &gc.Clicks, &gc.Reviews); err != nil {
			return s, err
		}
		s.TopGames = append(s.TopGames, gc)
	}
	if err := rows.Err(); err != nil {
		return s, err
	}
	return s, nil
}

func (r *Repo) ListAllReviews() ([]model.AdminReview, error) {
	rows, err := r.db.Query(`
		SELECT rv.id, rv.translation_id, rv.author_name, rv.rating, rv.body, rv.created_at,
		       g.title, g.slug, t.studio_name
		FROM reviews rv
		JOIN translations t ON t.id = rv.translation_id
		JOIN games g ON g.id = t.game_id
		ORDER BY rv.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AdminReview
	for rows.Next() {
		var ar model.AdminReview
		if err := rows.Scan(&ar.ID, &ar.TranslationID, &ar.AuthorName, &ar.Rating,
			&ar.Body, &ar.CreatedAt, &ar.GameTitle, &ar.GameSlug, &ar.StudioName); err != nil {
			return nil, err
		}
		out = append(out, ar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) ListAllTranslations() ([]model.AdminTranslation, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.game_id, t.studio_name, t.translator_names, t.type,
		       COALESCE(t.orthography,'academic'), t.coverage, t.external_url,
		       t.click_count, t.created_at, t.updated_at,
		       g.title, g.slug
		FROM translations t
		JOIN games g ON g.id = t.game_id
		ORDER BY g.title, t.created_at DESC`)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		t         model.Translation
		namesJSON string
		covJSON   string
		gTitle    string
		gSlug     string
	}
	for rows.Next() {
		var x struct {
			t         model.Translation
			namesJSON string
			covJSON   string
			gTitle    string
			gSlug     string
		}
		if err := rows.Scan(&x.t.ID, &x.t.GameID, &x.t.StudioName, &x.namesJSON, &x.t.Type,
			&x.t.Orthography, &x.covJSON, &x.t.ExternalURL,
			&x.t.ClickCount, &x.t.CreatedAt, &x.t.UpdatedAt,
			&x.gTitle, &x.gSlug); err != nil {
			rows.Close()
			return nil, err
		}
		raw = append(raw, x)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	var out []model.AdminTranslation
	for _, x := range raw {
		json.Unmarshal([]byte(x.namesJSON), &x.t.TranslatorNames)
		json.Unmarshal([]byte(x.covJSON), &x.t.Coverage)
		at := model.AdminTranslation{
			TranslationDetail: model.TranslationDetail{Translation: x.t},
			GameTitle:         x.gTitle,
			GameSlug:          x.gSlug,
		}
		var avg float64
		r.db.QueryRow(`SELECT COUNT(*), COALESCE(AVG(rating),0) FROM reviews WHERE translation_id=?`,
			x.t.ID).Scan(&at.ReviewCount, &avg)
		if at.ReviewCount > 0 {
			at.AvgRating = &avg
		}
		out = append(out, at)
	}
	return out, nil
}

func (r *Repo) ListAllTranslationSubmissions() ([]model.AdminTranslationSubmission, error) {
	rows, err := r.db.Query(`
		SELECT id, game_title, platforms, category, localization_type, authors,
		       game_url, translation_url, description, status, created_at, updated_at
		FROM translation_submissions
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.AdminTranslationSubmission
	for rows.Next() {
		var s model.AdminTranslationSubmission
		var platformsJSON, typesJSON string
		if err := rows.Scan(&s.ID, &s.GameTitle, &platformsJSON, &s.Category, &typesJSON,
			&s.Authors, &s.GameURL, &s.TranslationURL, &s.Description, &s.Status,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(platformsJSON), &s.Platforms)
		json.Unmarshal([]byte(typesJSON), &s.LocalizationType)
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) AcceptTranslationSubmission(id int) error {
	_, err := r.db.Exec(`UPDATE translation_submissions SET status='accepted', updated_at=? WHERE id=?`, time.Now(), id)
	return err
}

func (r *Repo) DeleteTranslationSubmission(id int) error {
	_, err := r.db.Exec(`DELETE FROM translation_submissions WHERE id=?`, id)
	return err
}

func (r *Repo) CreateGenre(name, slug string) (int64, error) {
	res, err := r.db.Exec(`INSERT INTO genres (name,slug) VALUES (?,?)`, name, slug)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) DeleteGenre(id int) error {
	_, err := r.db.Exec(`DELETE FROM genres WHERE id=?`, id)
	return err
}

func (r *Repo) GetTranslationByID(id int) (*model.Translation, error) {
	var t model.Translation
	var namesJSON, covJSON string
	err := r.db.QueryRow(`
		SELECT id, game_id, studio_name, translator_names, type,
		       COALESCE(orthography,'academic'), coverage, external_url,
		       click_count, created_at, updated_at
		FROM translations WHERE id=?`, id).
		Scan(&t.ID, &t.GameID, &t.StudioName, &namesJSON, &t.Type,
			&t.Orthography, &covJSON, &t.ExternalURL,
			&t.ClickCount, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(namesJSON), &t.TranslatorNames)
	json.Unmarshal([]byte(covJSON), &t.Coverage)
	return &t, nil
}
