package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"javar/internal/db"
	"javar/internal/handler"
)

// loadEnv чытае .env файл і ўстанаўлівае зменныя асяроддзя.
// Не перазапісвае ўжо выстаўленыя зменныя (сістэмныя маюць прыярытэт).
func loadEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return // .env неабавязковы
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
func main() {
	loadEnv()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	adminToken := os.Getenv("ADMIN_TOKEN") // optional

	database, err := db.Open("javar.db")
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	h := handler.New(database)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors)

	// ── Public API ────────────────────────────────────────
	r.Route("/api", func(r chi.Router) {
		r.Get("/genres", h.ListGenres)

		r.Get("/games", h.ListGames)
		r.Get("/games/{slug}", h.GetGame)

		r.Post("/translations/{id}/click", h.TrackClick)

		r.Get("/reviews", h.ListReviews)
		r.Post("/reviews", h.CreateReview)
	})

	// ── Admin API (token-protected) ───────────────────────
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(h.AdminMiddleware(adminToken))

		r.Post("/games", h.CreateGame)
		r.Put("/games/{id}", h.UpdateGame)
		r.Delete("/games/{id}", h.DeleteGame)

		r.Post("/translations", h.CreateTranslation)
		r.Put("/translations/{id}", h.UpdateTranslation)
		r.Delete("/translations/{id}", h.DeleteTranslation)

		r.Delete("/reviews/{id}", h.DeleteReview)

		r.Get("/stats", h.GetStats)
		r.Get("/reviews", h.ListAllReviews)
		r.Get("/translations", h.ListAllTranslations)
		r.Post("/genres", h.CreateGenre)
		r.Delete("/genres/{id}", h.DeleteGenre)
		r.Get("/export/csv", h.ExportCSV)
		r.Post("/upload", h.UploadImage)
	})

	// ── Static frontend ────────────────────────────────────
	r.Handle("/*", http.FileServer(http.Dir("./frontend")))

	log.Printf("▶  http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
