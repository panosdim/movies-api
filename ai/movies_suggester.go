package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"movies-backend/models"
	"movies-backend/utils"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/patrickmn/go-cache"
	"google.golang.org/api/option"
)

type PopularMovie struct {
	PosterPath  string `json:"poster_path"`
	ReleaseDate string `json:"release_date"`
	ID          int64  `json:"id"`
	Title       string `json:"title"`
}

func GetMoviesSuggestion(uid uint, numMovies int) ([]PopularMovie, error) {
	// Create a cache key based on user ID and number of movies
	cacheKey := fmt.Sprintf("moviesuggestion-%d-%d", uid, numMovies)
	// Check if the response is already cached
	if cachedResponse, found := utils.MovieSuggestionCache.Get(cacheKey); found {
		return cachedResponse.([]PopularMovie), nil
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Using `ResponseMIMEType` requires either a Gemini 1.5 Pro or 1.5 Flash model
	model := client.GenerativeModel(os.Getenv("AI_MODEL"))

	model.SetTemperature(1)
	model.SetTopK(64)
	model.SetTopP(0.95)
	model.ResponseMIMEType = "text/plain"
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are a movie expert.")},
	}

	// Get current list of popular movies
	options := map[string]string{
		"language": "en-US",
		"page":     "1",
	}

	page1, err := models.TMDbClient.GetMoviePopular(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular movies: %w", err)
	}

	options["page"] = "2"
	page2, err := models.TMDbClient.GetMoviePopular(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular movies: %w", err)
	}

	// Combine results from both pages
	popularMovies := append(page1.Results, page2.Results...)

	// Convert popularMovies to a list of objects with only id, title, release_date, and poster_path
	var simplifiedMovies []PopularMovie
	for _, movie := range popularMovies {
		simplifiedMovies = append(simplifiedMovies, PopularMovie{
			ID:          movie.ID,
			Title:       movie.Title,
			ReleaseDate: movie.ReleaseDate,
			PosterPath:  movie.PosterPath,
		})
	}

	// Convert simplifiedMovies to a string representation
	popularMoviesStr, err := json.Marshal(simplifiedMovies)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal popular movies: %w", err)
	}

	// Get current watchlist
	wl, err := models.GetWatchlistByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	// Convert watchlist to a list of objects with only title, movie_id, and rating
	var simplifiedWatchlist []struct {
		Title   string `json:"title"`
		MovieID uint   `json:"movie_id"`
		Rating  uint   `json:"rating"`
	}
	for _, movie := range wl {
		simplifiedWatchlist = append(simplifiedWatchlist, struct {
			Title   string `json:"title"`
			MovieID uint   `json:"movie_id"`
			Rating  uint   `json:"rating"`
		}{
			Title:   movie.Title,
			MovieID: movie.MovieID,
			Rating:  movie.Rating,
		})
	}

	// Convert simplifiedWatchlist to a string representation
	watchListStr, err := json.Marshal(simplifiedWatchlist)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal watchlist: %w", err)
	}

	// Get already seen movies
	watched, err := models.GetMoviesByUserID(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get already seen movies: %w", err)
	}

	// Convert already seen movies to a list of objects with only title, movie_id, and rating
	var simplifiedWatched []struct {
		Title   string `json:"title"`
		MovieID uint   `json:"movie_id"`
		Rating  uint   `json:"rating"`
	}
	for _, movie := range watched {
		simplifiedWatched = append(simplifiedWatched, struct {
			Title   string `json:"title"`
			MovieID uint   `json:"movie_id"`
			Rating  uint   `json:"rating"`
		}{
			Title:   movie.Title,
			MovieID: movie.MovieID,
			Rating:  movie.Rating,
		})
	}

	// Convert simplifiedWatched to a string representation
	watchedStr, err := json.Marshal(simplifiedWatched)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal already seen movies: %w", err)
	}

	prompt := fmt.Sprintf(`
	Here is my current watchlist: %s
	Here is my already watched movies: %s
	The rating value has range 0 to 5 with 5 representing the best value, 1 representing the worst value and 0 representing that the movie is not rated yet.
	The already watched movies and the watchlist have a mapping between id and movie_id the the list of movies.
	Suggest me %d movies that I would like to see, not already in my watchlist or already watched from following list of movies: %s`, watchListStr, watchedStr, numMovies, popularMoviesStr)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	suggestions, err := utils.GetTextResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Create a prompt for formatting the JSON representation
	format_prompt := fmt.Sprintf(`%s \n Format the above list of movies in JSON representation with properties title and id`, suggestions)

	// Generate formatted content
	format_resp, err := model.GenerateContent(ctx, genai.Text(format_prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	moviesList, err := utils.GetTextResponse(format_resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	var movieResults []PopularMovie

	// Create a map to store popular movies by their IDs for quick lookup
	popularMoviesMap := make(map[int64]PopularMovie)
	for _, movie := range simplifiedMovies {
		// Convert each movie to a PopularMovie struct
		popularMoviesMap[movie.ID] = PopularMovie{
			ID:          movie.ID,
			Title:       movie.Title,
			ReleaseDate: movie.ReleaseDate,
			PosterPath:  movie.PosterPath,
		}
	}

	// Format and clean the JSON content
	cleanedJsonStr := strings.TrimSpace(moviesList)
	cleanedJsonStr = strings.TrimPrefix(cleanedJsonStr, "```json")
	cleanedJsonStr = strings.TrimPrefix(cleanedJsonStr, "```json\n")
	cleanedJsonStr = strings.TrimSuffix(cleanedJsonStr, "\n```")
	cleanedJsonStr = strings.TrimSuffix(cleanedJsonStr, "```")

	// Define a struct to match the JSON structure
	type Movie struct {
		Title string `json:"title"`
		ID    int    `json:"id"`
	}

	// Unmarshal the JSON data into a slice of Movie structs
	var movies []Movie
	err = json.Unmarshal([]byte(cleanedJsonStr), &movies)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	// Populate movieResults using movies and popularMoviesMap
	for _, m := range movies {
		if movie, ok := popularMoviesMap[int64(m.ID)]; ok {
			movieResults = append(movieResults, movie)
		} else {
			return nil, fmt.Errorf("failed to find movie id: %d", m.ID)
		}
	}

	// Cache the response for future requests
	utils.MovieSuggestionCache.Set(cacheKey, movieResults, cache.DefaultExpiration)

	return movieResults, nil
}
