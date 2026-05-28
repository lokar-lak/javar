package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"javar/internal/model"
	"javar/internal/repository"
)

const adminSessionCookie = "javar_admin_session"

type Handler struct {
	repo *repository.Repo
}

func New(db *sql.DB) *Handler {
	return &Handler{repo: repository.New(db)}
}

// ── helpers ───────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ── GET /api/genres ───────────────────────────────────────────────────────

func (h *Handler) ListGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.repo.ListGenres()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if genres == nil {
		genres = []model.Genre{}
	}
	writeJSON(w, 200, genres)
}

// ── GET /api/translators ─────────────────────────────────────────────────

func (h *Handler) ListTranslators(w http.ResponseWriter, r *http.Request) {
	translators, err := h.repo.ListTranslators()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if translators == nil {
		translators = []string{}
	}
	writeJSON(w, 200, translators)
}

// ── GET /api/stats ────────────────────────────────────────────────────────

func (h *Handler) GetPublicStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetPublicStats()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 200, stats)
}

// ── GET /api/games ────────────────────────────────────────────────────────
// ?search=  &genre_id=  &type=manual|ai  &orthography=  &official_status=official|semi-official|unofficial
// &sort_by=created_at|release_date|steam_rating|best_rating  &sort_order=desc  &page=  &limit=

func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	genreID, _ := strconv.Atoi(q.Get("genre_id"))
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	games, err := h.repo.ListGames(model.GameFilter{
		Search:      q.Get("search"),
		GenreID:     genreID,
		Type:        q.Get("type"),
		Orthography: q.Get("orthography"),
		Official:    q.Get("official_status"),
		Translator:  q.Get("translator"),
		SortBy:      q.Get("sort_by"),
		SortOrder:   q.Get("sort_order"),
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if games == nil {
		games = []model.Game{}
	}
	total, err := h.repo.CountGames(model.GameFilter{
		Search:      q.Get("search"),
		GenreID:     genreID,
		Type:        q.Get("type"),
		Orthography: q.Get("orthography"),
		Official:    q.Get("official_status"),
		Translator:  q.Get("translator"),
	})
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	enrichGamesFromSteam(r.Context(), games)
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	writeJSON(w, 200, games)
}

// ── GET /api/games/{slug} ─────────────────────────────────────────────────

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	game, err := h.repo.GetGameBySlug(slug)
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if game == nil {
		writeError(w, 404, "game not found")
		return
	}
	enrichGameFromSteam(r.Context(), &game.Game)
	writeJSON(w, 200, game)
}

// ── POST /api/translations/{id}/click ─────────────────────────────────────

func (h *Handler) TrackClick(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 400, "invalid id")
		return
	}
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if parts := strings.SplitN(forwarded, ",", 2); len(parts) > 0 {
			ip = strings.TrimSpace(parts[0])
		}
	}
	url, err := h.repo.IncrementClick(id, ip)
	if err != nil {
		writeError(w, 404, "translation not found")
		return
	}
	writeJSON(w, 200, map[string]string{"url": url})
}

// ── GET /api/reviews?translation_id=X ────────────────────────────────────

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.URL.Query().Get("translation_id"))
	if err != nil || tid <= 0 {
		writeError(w, 400, "translation_id required")
		return
	}
	reviews, err := h.repo.ReviewsByTranslation(tid)
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	writeJSON(w, 200, reviews)
}

// ── POST /api/reviews ─────────────────────────────────────────────────────
// Body: { translation_id, author_name?, rating, body }

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	if req.AuthorName == "" {
		req.AuthorName = "Ананім"
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeError(w, 400, "rating must be 1–5")
		return
	}
	if req.Body == "" {
		writeError(w, 400, "body required")
		return
	}

	id, err := h.repo.CreateReview(req)
	if err != nil {
		// Duplicate review
		writeError(w, 409, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}

// ── POST /api/translation-submissions ─────────────────────────────────────

func (h *Handler) CreateTranslationSubmission(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTranslationSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	req.GameTitle = strings.TrimSpace(req.GameTitle)
	req.Authors = strings.TrimSpace(req.Authors)
	req.GameURL = strings.TrimSpace(req.GameURL)
	req.TranslationURL = strings.TrimSpace(req.TranslationURL)
	req.Description = strings.TrimSpace(req.Description)

	if req.GameTitle == "" || len(req.Platforms) == 0 || req.Category == "" || len(req.LocalizationType) == 0 || req.GameURL == "" {
		writeError(w, 400, "all fields are required")
		return
	}
	if req.Category != "official" && req.Category != "unofficial" {
		writeError(w, 400, "invalid category")
		return
	}
	if req.Category == "unofficial" && req.Authors == "" {
		writeError(w, 400, "authors required")
		return
	}
	if req.Category == "official" {
		req.Authors = ""
		req.TranslationURL = ""
	}
	if !validHTTPURL(req.GameURL) || (req.TranslationURL != "" && !validHTTPURL(req.TranslationURL)) {
		writeError(w, 400, "links must start with http:// or https://")
		return
	}

	similar, err := h.repo.FindSimilarGames(req.GameTitle)
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if len(similar) > 0 {
		writeJSON(w, 409, map[string]any{
			"error":         "similar game exists",
			"similar_games": similar,
		})
		return
	}

	id, err := h.repo.CreateTranslationSubmission(req)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}

func validHTTPURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (h *Handler) ListAllGames(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	games, err := h.repo.ListAllGames(search)
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if games == nil {
		games = []model.Game{}
	}
	writeJSON(w, 200, games)
}

// ── ADMIN ─────────────────────────────────────────────────────────────────

func (h *Handler) AdminLogin(passwordHash, sessionSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if passwordHash == "" || sessionSecret == "" {
			writeError(w, 503, "admin auth is not configured")
			return
		}

		var req struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
			writeError(w, 400, "password required")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeError(w, 401, "unauthorized")
			return
		}

		setAdminSessionCookie(w, r, sessionSecret, time.Now().Add(8*time.Hour))
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}
}

func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   adminCookieSecure(r),
	})
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) AdminMiddleware(sessionSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sessionSecret == "" || !validAdminSession(r, sessionSecret) {
				writeError(w, 401, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func setAdminSessionCookie(w http.ResponseWriter, r *http.Request, secret string, expires time.Time) {
	exp := strconv.FormatInt(expires.Unix(), 10)
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    exp + "." + signAdminSession(exp, secret),
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   adminCookieSecure(r),
	})
}

func validAdminSession(r *http.Request, secret string) bool {
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return false
	}
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return false
	}
	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	return hmac.Equal([]byte(parts[1]), []byte(signAdminSession(parts[0], secret)))
}

func signAdminSession(value, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func adminCookieSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// POST /api/admin/games
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	if req.Title == "" || req.Slug == "" || req.Developer == "" {
		writeError(w, 400, "title, slug, developer required")
		return
	}
	id, err := h.repo.CreateGame(req)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}

// PUT /api/admin/games/{id}
func (h *Handler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	if err := h.repo.UpdateGame(id, req); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// DELETE /api/admin/games/{id}
func (h *Handler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteGame(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// GET /api/admin/games/{id}/relations
func (h *Handler) GameRelations(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	translations, reviews := h.repo.GameRelationsCount(id)
	writeJSON(w, 200, map[string]int{"translations": translations, "reviews": reviews})
}

// POST /api/admin/translations
func (h *Handler) CreateTranslation(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	id, err := h.repo.CreateTranslation(req)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}

// PUT /api/admin/translations/{id}
func (h *Handler) UpdateTranslation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.CreateTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	if err := h.repo.UpdateTranslation(id, req); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// DELETE /api/admin/translations/{id}
func (h *Handler) DeleteTranslation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteTranslation(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// GET /api/admin/translations/{id}/relations
func (h *Handler) TranslationRelations(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	reviews := h.repo.TranslationRelationsCount(id)
	writeJSON(w, 200, map[string]int{"reviews": reviews})
}

// DELETE /api/admin/reviews/{id}
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteReview(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// ── GET /api/admin/steam-meta?url=... ─────────────────────────────────────

func (h *Handler) GetSteamMeta(w http.ResponseWriter, r *http.Request) {
	steamURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if steamURL == "" {
		writeError(w, 400, "url required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()
	meta, err := fetchSteamGameMeta(ctx, steamURL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, meta)
}

// ── GET /api/admin/stats ──────────────────────────────────────────────────

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetStats()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 200, stats)
}

// ── GET /api/admin/reviews ────────────────────────────────────────────────

func (h *Handler) ListAllReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.repo.ListAllReviews()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if reviews == nil {
		reviews = []model.AdminReview{}
	}
	writeJSON(w, 200, reviews)
}

// ── GET /api/admin/translations ───────────────────────────────────────────

func (h *Handler) ListAllTranslations(w http.ResponseWriter, r *http.Request) {
	tr, err := h.repo.ListAllTranslations()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if tr == nil {
		tr = []model.AdminTranslation{}
	}
	writeJSON(w, 200, tr)
}

// ── GET /api/admin/translation-submissions ───────────────────────────────

func (h *Handler) ListAllTranslationSubmissions(w http.ResponseWriter, r *http.Request) {
	submissions, err := h.repo.ListAllTranslationSubmissions()
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	if submissions == nil {
		submissions = []model.AdminTranslationSubmission{}
	}
	writeJSON(w, 200, submissions)
}

// ── POST /api/admin/translation-submissions/{id}/accept ──────────────────

func (h *Handler) AcceptTranslationSubmission(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.AcceptTranslationSubmission(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "accepted"})
}

// ── DELETE /api/admin/translation-submissions/{id} ───────────────────────

func (h *Handler) DeleteTranslationSubmission(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteTranslationSubmission(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// ── POST /api/admin/genres ────────────────────────────────────────────────

func (h *Handler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Slug == "" {
		writeError(w, 400, "name and slug required")
		return
	}
	id, err := h.repo.CreateGenre(req.Name, req.Slug)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}

// ── DELETE /api/admin/genres/{id} ────────────────────────────────────────

func (h *Handler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteGenre(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// ── GET /api/admin/export/csv ─────────────────────────────────────────────

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.repo.GetStats()
	reviews, _ := h.repo.ListAllReviews()

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="javar_export.csv"`)

	fmt.Fprintf(w, "# Статыстыка\n")
	fmt.Fprintf(w, "Гульняў,%d\nПеракладаў,%d\nВодгукаў,%d\nКлікаў,%d\n\n",
		stats.TotalGames, stats.TotalTranslations, stats.TotalReviews, stats.TotalClicks)

	fmt.Fprintf(w, "# Топ гульняў па кліках\n")
	fmt.Fprintf(w, "Назва,Клікаў,Водгукаў\n")
	for _, g := range stats.TopGames {
		fmt.Fprintf(w, "%q,%d,%d\n", g.Title, g.Clicks, g.Reviews)
	}

	fmt.Fprintf(w, "\n# Водгукі\n")
	fmt.Fprintf(w, "Гульня,Пераклад,Аўтар,Рэйтынг,Тэкст,Дата\n")
	for _, rv := range reviews {
		translationName := strings.Join(rv.TranslatorNames, ", ")
		if translationName == "" {
			translationName = "Беларусізатар"
		}
		fmt.Fprintf(w, "%q,%q,%q,%d,%q,%s\n",
			rv.GameTitle, translationName, rv.AuthorName,
			rv.Rating, rv.Body, rv.CreatedAt.Format("2006-01-02"))
	}
}

// ── POST /api/admin/upload ────────────────────────────────────────────────
// Прымае multipart/form-data з полем "file", захоўвае ў frontend/uploads/

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // максімум 10 MB

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "file required")
		return
	}
	defer file.Close()

	// Толькі выявы
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		writeError(w, 400, "only jpg, png, webp allowed")
		return
	}

	// Унікальнае імя файла
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join("frontend", "uploads", name)

	if err := os.MkdirAll(filepath.Join("frontend", "uploads"), 0755); err != nil {
		writeError(w, 500, "cannot create uploads dir")
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		writeError(w, 500, "cannot save file")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		writeError(w, 500, "cannot write file")
		return
	}

	writeJSON(w, 200, map[string]string{"url": "/uploads/" + name})
}
