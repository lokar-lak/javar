package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"javar/internal/model"
)

var steamAppIDRx = regexp.MustCompile(`/app/(\d+)`)

// enrichGameFromSteam fetches ONLY cover image and rating from Steam API.
// Description is not touched; it is stored manually in the DB.
func enrichGameFromSteam(ctx context.Context, game *model.Game) {
	if game == nil || game.SteamDBURL == "" {
		return
	}
	appID, ok := steamAppIDFromURL(game.SteamDBURL)
	if !ok {
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Cover image: if DB already has a URL (including manual one), do not override it.
	// Fallback to Steam only when cover_url is effectively empty.
	if strings.TrimSpace(game.CoverURL) == "" {
		if img, err := fetchSteamHeaderImage(reqCtx, appID); err == nil && img != "" {
			game.CoverURL = img
		}
	}

	// Rating (only when not set manually)
	if game.SteamRating == nil {
		if rating, err := fetchSteamRating(reqCtx, appID); err == nil {
			game.SteamRating = rating
		}
	}
}

func steamAppIDFromURL(u string) (int, bool) {
	m := steamAppIDRx.FindStringSubmatch(u)
	if len(m) < 2 {
		return 0, false
	}
	id, err := strconv.Atoi(m[1])
	return id, err == nil && id > 0
}

func fetchSteamHeaderImage(ctx context.Context, appID int) (string, error) {
	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&l=english", appID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]struct {
		Success bool `json:"success"`
		Data    struct {
			HeaderImage string `json:"header_image"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	key := strconv.Itoa(appID)
	entry, ok := data[key]
	if !ok || !entry.Success {
		return "", fmt.Errorf("app %d not found", appID)
	}
	return entry.Data.HeaderImage, nil
}

func fetchSteamGameMeta(ctx context.Context, appURL string) (*model.SteamGameMeta, error) {
	appID, ok := steamAppIDFromURL(appURL)
	if !ok {
		return nil, fmt.Errorf("invalid steam app url")
	}

	var meta model.SteamGameMeta
	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&l=english", appID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]struct {
		Success bool `json:"success"`
		Data    struct {
			Name        string   `json:"name"`
			Developers  []string `json:"developers"`
			Publishers  []string `json:"publishers"`
			HeaderImage string   `json:"header_image"`
			ReleaseDate struct {
				Date string `json:"date"`
			} `json:"release_date"`
			Genres []struct {
				Description string `json:"description"`
			} `json:"genres"`
			Platforms struct {
				Windows bool `json:"windows"`
				Mac     bool `json:"mac"`
				Linux   bool `json:"linux"`
			} `json:"platforms"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	entry, ok := data[strconv.Itoa(appID)]
	if !ok || !entry.Success {
		return nil, fmt.Errorf("app %d not found", appID)
	}

	meta.Title = entry.Data.Name
	if len(entry.Data.Developers) > 0 {
		meta.Developer = strings.Join(entry.Data.Developers, ", ")
	}
	if len(entry.Data.Publishers) > 0 {
		meta.Publisher = strings.Join(entry.Data.Publishers, ", ")
	}
	meta.CoverURL = entry.Data.HeaderImage
	meta.ReleaseDate = steamReleaseDateToISO(entry.Data.ReleaseDate.Date)
	for _, g := range entry.Data.Genres {
		if g.Description != "" {
			meta.Genres = append(meta.Genres, g.Description)
		}
	}
	if entry.Data.Platforms.Windows {
		meta.Platforms = append(meta.Platforms, "ПК")
	}
	if entry.Data.Platforms.Mac {
		meta.Platforms = append(meta.Platforms, "macOS")
	}
	if entry.Data.Platforms.Linux {
		meta.Platforms = append(meta.Platforms, "Linux")
	}
	if rating, err := fetchSteamRating(ctx, appID); err == nil {
		meta.SteamRating = rating
	}
	return &meta, nil
}

func steamReleaseDateToISO(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return ""
	}
	for _, layout := range []string{"2 Jan, 2006", "Jan 2, 2006", "2 Jan 2006", "Jan 2006", "2006"} {
		if t, err := time.Parse(layout, date); err == nil {
			if layout == "Jan 2006" || layout == "2006" {
				return t.Format("2006-01")
			}
			return t.Format("2006-01-02")
		}
	}
	return ""
}

func fetchSteamRating(ctx context.Context, appID int) (*int, error) {
	url := fmt.Sprintf(
		"https://store.steampowered.com/appreviews/%d?json=1&language=all&purchase_type=all&num_per_page=0",
		appID,
	)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Success      int `json:"success"`
		QuerySummary struct {
			TotalPositive int `json:"total_positive"`
			TotalNegative int `json:"total_negative"`
		} `json:"query_summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	total := payload.QuerySummary.TotalPositive + payload.QuerySummary.TotalNegative
	if total <= 0 {
		return nil, nil
	}
	score := int(float64(payload.QuerySummary.TotalPositive) / float64(total) * 100)
	return &score, nil
}
