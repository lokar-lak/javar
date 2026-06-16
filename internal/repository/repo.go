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

var searchReplacements = [][2]string{
	{"ą", "a"}, {"Ą", "a"}, {"á", "a"}, {"Á", "a"}, {"à", "a"}, {"À", "a"}, {"â", "a"}, {"Â", "a"}, {"ä", "a"}, {"Ä", "a"}, {"ã", "a"}, {"Ã", "a"}, {"å", "a"}, {"Å", "a"}, {"ā", "a"}, {"Ā", "a"}, {"ă", "a"}, {"Ă", "a"},
	{"ć", "c"}, {"Ć", "c"}, {"č", "c"}, {"Č", "c"}, {"ç", "c"}, {"Ç", "c"},
	{"ę", "e"}, {"Ę", "e"}, {"é", "e"}, {"É", "e"}, {"è", "e"}, {"È", "e"}, {"ê", "e"}, {"Ê", "e"}, {"ë", "e"}, {"Ë", "e"}, {"ē", "e"}, {"Ē", "e"}, {"ě", "e"}, {"Ě", "e"},
	{"í", "i"}, {"Í", "i"}, {"ì", "i"}, {"Ì", "i"}, {"î", "i"}, {"Î", "i"}, {"ï", "i"}, {"Ï", "i"}, {"ī", "i"}, {"Ī", "i"},
	{"ł", "l"}, {"Ł", "l"},
	{"ń", "n"}, {"Ń", "n"}, {"ñ", "n"}, {"Ñ", "n"},
	{"ó", "o"}, {"Ó", "o"}, {"ò", "o"}, {"Ò", "o"}, {"ô", "o"}, {"Ô", "o"}, {"ö", "o"}, {"Ö", "o"}, {"õ", "o"}, {"Õ", "o"}, {"ø", "o"}, {"Ø", "o"}, {"ō", "o"}, {"Ō", "o"},
	{"ś", "s"}, {"Ś", "s"}, {"š", "s"}, {"Š", "s"}, {"ş", "s"}, {"Ş", "s"},
	{"ú", "u"}, {"Ú", "u"}, {"ù", "u"}, {"Ù", "u"}, {"û", "u"}, {"Û", "u"}, {"ü", "u"}, {"Ü", "u"}, {"ū", "u"}, {"Ū", "u"},
	{"ý", "y"}, {"Ý", "y"}, {"ÿ", "y"}, {"Ÿ", "y"},
	{"ź", "z"}, {"Ź", "z"}, {"ż", "z"}, {"Ż", "z"}, {"ž", "z"}, {"Ž", "z"},
}

var sqlSearchReplacements = [][2]string{
	{"ą", "a"}, {"Ą", "a"}, {"ć", "c"}, {"Ć", "c"}, {"č", "c"}, {"Č", "c"},
	{"ę", "e"}, {"Ę", "e"}, {"ł", "l"}, {"Ł", "l"}, {"ń", "n"}, {"Ń", "n"},
	{"ó", "o"}, {"Ó", "o"}, {"ś", "s"}, {"Ś", "s"}, {"š", "s"}, {"Š", "s"},
	{"ź", "z"}, {"Ź", "z"}, {"ż", "z"}, {"Ż", "z"}, {"ž", "z"}, {"Ž", "z"},
}

func normalizeSearchText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, repl := range searchReplacements {
		value = strings.ReplaceAll(value, repl[0], repl[1])
	}
	return value
}

func normalizedSQL(expr string) string {
	out := "lower(" + expr + ")"
	for _, repl := range sqlSearchReplacements {
		out = "replace(" + out + ", '" + repl[0] + "', '" + repl[1] + "')"
	}
	return out
}

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

func (r *Repo) ListTranslators() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT name FROM (
			SELECT trim(j.value) AS name
			FROM translations t, json_each(t.translator_names) j
			WHERE trim(j.value) <> ''
			UNION
			SELECT trim(studio) AS name
			FROM translations
			WHERE trim(studio) <> ''
		)
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ═══ GAMES ═══════════════════════════════════════════════════════════════

func (r *Repo) ListGames(f model.GameFilter) ([]model.Game, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 20
	}

	q := `SELECT DISTINCT g.id, g.title, g.slug, g.developer,
	       COALESCE(g.publisher,''), COALESCE(g.release_date,''),
	       COALESCE(g.description,''), COALESCE(g.cover_url,''), COALESCE(g.steamdb_url,''),
	       COALESCE(g.platforms,'[]'), g.steam_rating, g.created_at,
	       COALESCE(
	         (SELECT MAX(avg_r) FROM (
	           SELECT AVG(rv.rating) as avg_r
	           FROM translations t2
	           LEFT JOIN reviews rv ON rv.translation_id = t2.id
	           WHERE t2.game_id = g.id
	           GROUP BY t2.id
	         )),
	         0
	       ) as best_rating,
	       (SELECT COUNT(*) FROM translations WHERE game_id = g.id) as translation_count
		FROM games g
		LEFT JOIN game_genres gg ON gg.game_id = g.id
		LEFT JOIN translations t  ON t.game_id  = g.id
		WHERE 1=1`

	var args []any

	if f.Search != "" {
		q += ` AND (` + normalizedSQL("g.title") + ` LIKE ? OR ` + normalizedSQL("g.developer") + ` LIKE ?)`
		like := "%" + normalizeSearchText(f.Search) + "%"
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
		q += ` AND t.orthography LIKE '%"' || ? || '"%'`
		args = append(args, f.Orthography)
	}
	if f.Official != "" {
		q += ` AND t.official_status = ?`
		args = append(args, f.Official)
	}
	if f.Translator != "" {
		q += ` AND (trim(t.studio) = ? OR EXISTS (SELECT 1 FROM json_each(t.translator_names) tn WHERE trim(tn.value) = ?))`
		args = append(args, f.Translator, f.Translator)
	}

	// ── Sorting ──────────────────────────────────────────────────────────
	sortBy := "COALESCE((SELECT MAX(t2.created_at) FROM translations t2 WHERE t2.game_id = g.id), g.created_at)"
	sortOrder := "DESC"
	sortMissing := ""
	switch f.SortBy {
	case "release_date":
		sortBy = `CASE
			WHEN g.release_date LIKE '__-__-____' THEN substr(g.release_date, 7, 4) || '-' || substr(g.release_date, 4, 2) || '-' || substr(g.release_date, 1, 2)
			WHEN g.release_date LIKE '__-____' THEN substr(g.release_date, 4, 4) || '-' || substr(g.release_date, 1, 2) || '-00'
			ELSE g.release_date
		END`
		sortMissing = `(g.release_date IS NULL OR g.release_date = '') ASC, `
	case "steam_rating":
		sortBy = "g.steam_rating"
	case "best_rating":
		sortBy = "best_rating"
	}
	if f.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	q += ` ORDER BY ` + sortMissing + sortBy + ` ` + sortOrder + ` NULLS LAST, g.id DESC LIMIT ? OFFSET ?`
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
		var platformsJSON string
		if err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Developer, &g.Publisher,
			&g.ReleaseDate, &g.Description, &g.CoverURL, &g.SteamDBURL, &platformsJSON, &rating, &g.CreatedAt, &g.BestRating,
			&g.TranslationCount); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(platformsJSON), &g.Platforms)
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
		games[i].HasOfficial = r.hasOfficial(games[i].ID)
		games[i].HasVerified = r.hasVerified(games[i].ID)
	}
	return games, nil
}

func (r *Repo) CountGames(f model.GameFilter) (int, error) {
	q := `SELECT COUNT(DISTINCT g.id)
		FROM games g
		LEFT JOIN game_genres gg ON gg.game_id = g.id
		LEFT JOIN translations t  ON t.game_id  = g.id
		WHERE 1=1`

	var args []any

	if f.Search != "" {
		q += ` AND (` + normalizedSQL("g.title") + ` LIKE ? OR ` + normalizedSQL("g.developer") + ` LIKE ?)`
		like := "%" + normalizeSearchText(f.Search) + "%"
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
		q += ` AND t.orthography LIKE '%"' || ? || '"%'`
		args = append(args, f.Orthography)
	}
	if f.Official != "" {
		q += ` AND t.official_status = ?`
		args = append(args, f.Official)
	}
	if f.Translator != "" {
		q += ` AND (trim(t.studio) = ? OR EXISTS (SELECT 1 FROM json_each(t.translator_names) tn WHERE trim(tn.value) = ?))`
		args = append(args, f.Translator, f.Translator)
	}

	var total int
	if err := r.db.QueryRow(q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repo) GetPublicStats() (*model.PublicStats, error) {
	s := &model.PublicStats{}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE official_status='official'`).Scan(&s.Official); err != nil {
		return nil, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE official_status='semi-official'`).Scan(&s.SemiOfficial); err != nil {
		return nil, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE official_status='unofficial'`).Scan(&s.Unofficial); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) GetGameBySlug(slug string) (*model.GameDetail, error) {
	var g model.Game
	var rating sql.NullInt64
	var platformsJSON string
	err := r.db.QueryRow(`
		SELECT id, title, slug, developer,
		       COALESCE(publisher,''), COALESCE(release_date,''),
		       COALESCE(description,''), COALESCE(cover_url,''),
		       COALESCE(steamdb_url,''), COALESCE(platforms,'[]'), steam_rating, created_at
		FROM games WHERE slug = ?`, slug).Scan(
		&g.ID, &g.Title, &g.Slug, &g.Developer, &g.Publisher,
		&g.ReleaseDate, &g.Description, &g.CoverURL, &g.SteamDBURL,
		&platformsJSON, &rating, &g.CreatedAt)
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
	json.Unmarshal([]byte(platformsJSON), &g.Platforms)
	g.Genres, _ = r.genresByGame(g.ID)
	g.HasOnlyAI = r.hasOnlyAI(g.ID)
	g.HasOfficial = r.hasOfficial(g.ID)
	g.HasVerified = r.hasVerified(g.ID)

	translations, err := r.translationsByGame(g.ID)
	if err != nil {
		return nil, err
	}
	return &model.GameDetail{Game: g, Translations: translations}, nil
}

func (r *Repo) CreateGame(req model.CreateGameRequest) (int64, error) {
	if req.Platforms == nil {
		req.Platforms = []string{}
	}
	platformsJSON, _ := json.Marshal(req.Platforms)
	res, err := r.db.Exec(`
		INSERT INTO games (title,slug,developer,publisher,release_date,description,cover_url,steamdb_url,platforms,steam_rating)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		req.Title, req.Slug, req.Developer, req.Publisher,
		req.ReleaseDate, req.Description, req.CoverURL, req.SteamDBURL, string(platformsJSON), req.SteamRating)
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
	if req.Platforms == nil {
		req.Platforms = []string{}
	}
	platformsJSON, _ := json.Marshal(req.Platforms)
	_, err := r.db.Exec(`
		UPDATE games SET title=?,slug=?,developer=?,publisher=?,
		  release_date=?,description=?,cover_url=?,steamdb_url=?,platforms=?,steam_rating=?
		WHERE id=?`,
		req.Title, req.Slug, req.Developer, req.Publisher,
		req.ReleaseDate, req.Description, req.CoverURL, req.SteamDBURL, string(platformsJSON), req.SteamRating, id)
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

func (r *Repo) GameRelationsCount(id int) (translations, reviews int) {
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=?`, id).Scan(&translations)
	r.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE translation_id IN (SELECT id FROM translations WHERE game_id=?)`, id).Scan(&reviews)
	return
}

func (r *Repo) hasOnlyAI(gameID int) bool {
	var total, manual int
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=?`, gameID).Scan(&total)
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=? AND type='manual'`, gameID).Scan(&manual)
	return total > 0 && manual == 0
}

func (r *Repo) hasOfficial(gameID int) bool {
	var cnt int
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=? AND official_status='official'`, gameID).Scan(&cnt)
	return cnt > 0
}

func (r *Repo) hasVerified(gameID int) bool {
	var cnt int
	r.db.QueryRow(`SELECT COUNT(*) FROM translations WHERE game_id=? AND verified=1`, gameID).Scan(&cnt)
	return cnt > 0
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
		SELECT id, game_id, COALESCE(game_title,''), COALESCE(studio,''), translator_names, type,
		       COALESCE(official_status,'unofficial'), COALESCE(orthography,'[]'),
		       coverage, external_url, verified, verified_at, incomplete, broken, click_count, created_at, updated_at
		FROM translations WHERE game_id=? ORDER BY created_at DESC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.TranslationDetail
	for rows.Next() {
		var t model.Translation
		var namesJSON, orthJSON, coverageJSON string
		var verifiedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.GameID, &t.GameTitle, &t.Studio, &namesJSON, &t.Type,
			&t.OfficialStatus, &orthJSON, &coverageJSON, &t.ExternalURL,
			&t.Verified, &verifiedAt, &t.Incomplete, &t.Broken, &t.ClickCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if verifiedAt.Valid {
			t.VerifiedAt = &verifiedAt.Time
		}
		json.Unmarshal([]byte(namesJSON), &t.TranslatorNames)
		json.Unmarshal([]byte(coverageJSON), &t.Coverage)
		json.Unmarshal([]byte(orthJSON), &t.Orthography)

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
	if req.OfficialStatus != "official" && req.OfficialStatus != "semi-official" && req.OfficialStatus != "unofficial" {
		req.OfficialStatus = "unofficial"
	}
	if len(req.Orthography) == 0 {
		req.Orthography = []string{"academic"}
	}
	namesJSON, _ := json.Marshal(req.TranslatorNames)
	req.Studio = strings.TrimSpace(req.Studio)
	req.Coverage = normalizeCoverageValues(req.Coverage)
	coverageJSON, _ := json.Marshal(req.Coverage)
	orthJSON, _ := json.Marshal(req.Orthography)
	var verifiedAt any
	if req.Verified {
		verifiedAt = time.Now()
	}
	res, err := r.db.Exec(`
		INSERT INTO translations (game_id,game_title,studio,translator_names,type,official_status,orthography,coverage,external_url,verified,verified_at,incomplete,broken)
		VALUES (?,(SELECT title FROM games WHERE id=?),?,?,?,?,?,?,?,?,?,?,?)`,
		req.GameID, req.GameID, req.Studio, string(namesJSON),
		req.Type, req.OfficialStatus, string(orthJSON), string(coverageJSON), req.ExternalURL, req.Verified, verifiedAt, req.Incomplete, req.Broken)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) UpdateTranslation(id int, req model.CreateTranslationRequest) error {
	if len(req.Orthography) == 0 {
		req.Orthography = []string{"academic"}
	}
	if req.OfficialStatus != "official" && req.OfficialStatus != "semi-official" && req.OfficialStatus != "unofficial" {
		req.OfficialStatus = "unofficial"
	}
	namesJSON, _ := json.Marshal(req.TranslatorNames)
	req.Studio = strings.TrimSpace(req.Studio)
	req.Coverage = normalizeCoverageValues(req.Coverage)
	coverageJSON, _ := json.Marshal(req.Coverage)
	orthJSON, _ := json.Marshal(req.Orthography)
	var wasVerified bool
	if err := r.db.QueryRow(`SELECT verified FROM translations WHERE id=?`, id).Scan(&wasVerified); err != nil {
		return err
	}
	now := time.Now()
	verifiedAtSQL := `verified_at`
	if req.Verified != wasVerified {
		if req.Verified {
			verifiedAtSQL = `?`
		} else {
			verifiedAtSQL = `NULL`
		}
	}
	q := `UPDATE translations SET studio=?,translator_names=?,type=?,official_status=?,orthography=?,
		  coverage=?,external_url=?,verified=?,incomplete=?,broken=?,verified_at=` + verifiedAtSQL + `,updated_at=? WHERE id=?`
	args := []any{req.Studio, string(namesJSON), req.Type, req.OfficialStatus, string(orthJSON),
		string(coverageJSON), req.ExternalURL, req.Verified, req.Incomplete, req.Broken}
	if req.Verified != wasVerified && req.Verified {
		args = append(args, now)
	}
	args = append(args, now, id)
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *Repo) DeleteTranslation(id int) error {
	_, err := r.db.Exec(`DELETE FROM translations WHERE id=?`, id)
	return err
}

func (r *Repo) TranslationRelationsCount(id int) (reviews int) {
	r.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE translation_id=?`, id).Scan(&reviews)
	return
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

func (r *Repo) IncrementClick(id int, ip string) (string, error) {
	var url string
	if err := r.db.QueryRow(`SELECT external_url FROM translations WHERE id=?`, id).Scan(&url); err != nil {
		return "", fmt.Errorf("not found")
	}
	var exists int
	err := r.db.QueryRow(`SELECT 1 FROM click_events WHERE translation_id=? AND ip=? AND created_at > datetime('now', '-24 hours')`, id, ip).Scan(&exists)
	if err == sql.ErrNoRows {
		r.db.Exec(`INSERT INTO click_events(translation_id, ip) VALUES(?,?)`, id, ip)
		r.db.Exec(`UPDATE translations SET click_count=click_count+1 WHERE id=?`, id)
	}
	return url, nil
}

func (r *Repo) ListAllGames(search string) ([]model.Game, error) {
	q := `SELECT id, title, slug, developer,
	       COALESCE(publisher,''), COALESCE(release_date,''),
	       COALESCE(description,''), COALESCE(cover_url,''), COALESCE(steamdb_url,''),
	       COALESCE(platforms,'[]'), steam_rating, created_at,
	       (SELECT COUNT(*) FROM translations WHERE game_id = g.id)
	FROM games g`
	var args []any
	if search != "" {
		q += ` WHERE ` + normalizedSQL("title") + ` LIKE ? OR ` + normalizedSQL("developer") + ` LIKE ?`
		like := "%" + normalizeSearchText(search) + "%"
		args = append(args, like, like)
	}
	q += ` ORDER BY title`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		var rating sql.NullInt64
		var platformsJSON string
		if err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Developer, &g.Publisher,
			&g.ReleaseDate, &g.Description, &g.CoverURL, &g.SteamDBURL, &platformsJSON, &rating, &g.CreatedAt,
			&g.TranslationCount); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(platformsJSON), &g.Platforms)
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
		games[i].HasOfficial = r.hasOfficial(games[i].ID)
	}
	return games, nil
}

// ═══ REVIEWS ═════════════════════════════════════════════════════════════

func (r *Repo) ReviewsByTranslation(id int) ([]model.Review, error) {
	rows, err := r.db.Query(`
		SELECT id, translation_id, reviewer_id, author_name, rating, body, created_at
		FROM reviews WHERE translation_id=? ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.TranslationID, &rv.ReviewerID, &rv.AuthorName, &rv.Rating, &rv.Body, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) SaveReview(req model.CreateReviewRequest) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(`SELECT id FROM reviews WHERE translation_id=? AND reviewer_id=?`, req.TranslationID, req.ReviewerID).Scan(&id)
	if err == nil {
		_, err = r.db.Exec(`
			UPDATE reviews
			SET author_name=?, rating=?, body=?, created_at=CURRENT_TIMESTAMP
			WHERE id=?`,
			strings.TrimSpace(req.AuthorName), req.Rating, strings.TrimSpace(req.Body), id)
		return id, false, err
	}
	if err != sql.ErrNoRows {
		return 0, false, err
	}

	res, err := r.db.Exec(`
		INSERT INTO reviews (translation_id,reviewer_id,author_name,rating,body) VALUES (?,?,?,?,?)`,
		req.TranslationID, req.ReviewerID, strings.TrimSpace(req.AuthorName), req.Rating, strings.TrimSpace(req.Body))
	if err != nil {
		return 0, false, err
	}
	id, err = res.LastInsertId()
	return id, true, err
}

func (r *Repo) HasReviewByReviewer(translationID int, reviewerID string) (bool, error) {
	var id int
	err := r.db.QueryRow(`SELECT id FROM reviews WHERE translation_id=? AND reviewer_id=?`, translationID, reviewerID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *Repo) CountRecentReviewEvents(reviewerID, ip string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM review_events
		WHERE created_at > datetime('now', '-10 minutes')
		  AND (reviewer_id=? OR ip=?)`, reviewerID, ip).Scan(&count)
	return count, err
}

func (r *Repo) RecordReviewEvent(reviewerID, ip string) error {
	_, err := r.db.Exec(`INSERT INTO review_events (reviewer_id, ip) VALUES (?, ?)`, reviewerID, ip)
	return err
}

func (r *Repo) FindSimilarGames(title string) ([]model.SimilarGame, error) {
	title = strings.TrimSpace(title)
	if len([]rune(title)) < 3 {
		return nil, nil
	}
	rows, err := r.db.Query(`
		SELECT title, slug FROM games
		WHERE `+normalizedSQL("title")+` LIKE ? OR ? LIKE '%' || `+normalizedSQL("title")+` || '%'
		ORDER BY title LIMIT 5`, "%"+normalizeSearchText(title)+"%", normalizeSearchText(title))
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
		SELECT rv.id, rv.translation_id, rv.reviewer_id, rv.author_name, rv.rating, rv.body, rv.created_at,
		       g.title, g.slug, COALESCE(t.studio,''), t.translator_names
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
		var namesJSON string
		if err := rows.Scan(&ar.ID, &ar.TranslationID, &ar.ReviewerID, &ar.AuthorName, &ar.Rating,
			&ar.Body, &ar.CreatedAt, &ar.GameTitle, &ar.GameSlug, &ar.Studio, &namesJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(namesJSON), &ar.TranslatorNames)
		out = append(out, ar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) ListAllTranslations() ([]model.AdminTranslation, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.game_id, COALESCE(t.game_title,''), COALESCE(t.studio,''), t.translator_names, t.type,
		       COALESCE(t.official_status,'unofficial'), COALESCE(t.orthography,'[]'), t.coverage, t.external_url,
		       t.verified, t.verified_at, t.incomplete, t.broken, t.click_count, t.created_at, t.updated_at,
		       g.title, g.slug
		FROM translations t
		JOIN games g ON g.id = t.game_id
		ORDER BY g.title, t.created_at DESC`)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		t          model.Translation
		namesJSON  string
		orthJSON   string
		covJSON    string
		verifiedAt sql.NullTime
		gTitle     string
		gSlug      string
	}
	for rows.Next() {
		var x struct {
			t          model.Translation
			namesJSON  string
			orthJSON   string
			covJSON    string
			verifiedAt sql.NullTime
			gTitle     string
			gSlug      string
		}
		if err := rows.Scan(&x.t.ID, &x.t.GameID, &x.t.GameTitle, &x.t.Studio, &x.namesJSON, &x.t.Type,
			&x.t.OfficialStatus, &x.orthJSON, &x.covJSON, &x.t.ExternalURL,
			&x.t.Verified, &x.verifiedAt, &x.t.ClickCount, &x.t.CreatedAt, &x.t.UpdatedAt,
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
		json.Unmarshal([]byte(x.orthJSON), &x.t.Orthography)
		json.Unmarshal([]byte(x.covJSON), &x.t.Coverage)
		if x.verifiedAt.Valid {
			x.t.VerifiedAt = &x.verifiedAt.Time
		}
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
	var namesJSON, orthJSON, covJSON string
	var verifiedAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, game_id, COALESCE(game_title,''), COALESCE(studio,''), translator_names, type,
		       COALESCE(official_status,'unofficial'), COALESCE(orthography,'[]'), coverage, external_url,
		       verified, verified_at, click_count, created_at, updated_at
		FROM translations WHERE id=?`, id).
		Scan(&t.ID, &t.GameID, &t.GameTitle, &t.Studio, &namesJSON, &t.Type,
			&t.OfficialStatus, &orthJSON, &covJSON, &t.ExternalURL,
			&t.Verified, &verifiedAt, &t.ClickCount, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(namesJSON), &t.TranslatorNames)
	json.Unmarshal([]byte(orthJSON), &t.Orthography)
	json.Unmarshal([]byte(covJSON), &t.Coverage)
	if verifiedAt.Valid {
		t.VerifiedAt = &verifiedAt.Time
	}
	return &t, nil
}
