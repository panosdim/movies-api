package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"movies-backend/utils"
	"net/http"
	"os"

	"github.com/patrickmn/go-cache"
)

type Suggestion struct {
	PosterPath      string  `json:"poster_path"`
	ReleaseDate     string  `json:"release_date"`
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	PredictedRating float64 `json:"predicted_rating"`
}

func GetMoviesSuggestion(uid uint, numMovies int) ([]Suggestion, error) {
	// Create a cache key based on user ID and number of movies
	cacheKey := fmt.Sprintf("moviesuggestion-%d-%d", uid, numMovies)
	// Check if the response is already cached
	if cachedResponse, found := utils.MovieSuggestionCache.Get(cacheKey); found {
		return cachedResponse.([]Suggestion), nil
	}

	// Make GET request to suggestions API
	resp, err := http.Get(os.Getenv("MOVIES_ML_BASE_URL") + "/suggestions")
	if err != nil {
		return nil, fmt.Errorf("failed to call suggestions API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("suggestions API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Decode JSON response
	var suggestions []Suggestion
	if err := json.NewDecoder(resp.Body).Decode(&suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Cache the response for future requests
	utils.MovieSuggestionCache.Set(cacheKey, suggestions, cache.DefaultExpiration)

	return suggestions, nil
}
