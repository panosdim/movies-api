package utils

import (
	"fmt"
	"io"
	"log"
	"movies-backend/models"
	"movies-backend/utils/mail"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	YYYYMMDD = "2006-01-02"
)

func CheckForAvailableMovies() {
	var users []models.User

	if err := models.DB.Find(&users).Error; err != nil {
		log.Println("Warning: Cannot get users from DB")
		return
	}

	for _, user := range users {
		// Find all movies without release release_date for each user
		movies := models.GetMoviesWithoutReleaseDateByUserID(user.ID)

		// Update release dates
		for _, movie := range movies {
			movie.UpdateReleaseDate()
			_ = movie.UpdateMovie()
		}

		// Get the movies that are available and we have not send an email notification
		now := time.Now().UTC()
		var availableMovies []models.Movie
		models.DB.Where("release_date <= ? AND email_sent = ?", now.Format(YYYYMMDD), 0).Find(&availableMovies, "user_id = ?", user.ID)

		if len(availableMovies) > 0 {
			movieTitles := []string{}
			for _, movie := range availableMovies {
				movieTitles = append(movieTitles, movie.Title)
			}

			err := mail.SendMail(user.Email, movieTitles)
			if err == nil {
				for _, movie := range availableMovies {
					movie.EmailSent = true
					movie.UpdateMovie()
				}
			}
		}
	}
}

// Create a cache with a default expiration time of 24 hours, and purge every 12 hours
var MovieSuggestionCache = cache.New(30*24*time.Hour, 12*time.Hour)

// Function to clear the cache for a specific user
func ClearUserMovieSuggestionCache(uid uint) {
	// Iterate through all cache keys and delete keys related to the user
	for k := range MovieSuggestionCache.Items() {
		if strings.HasPrefix(k, fmt.Sprintf("moviesuggestion-%d-", uid)) {
			MovieSuggestionCache.Delete(k)
		}
	}
}

func TriggerModelRetrain() {
	// Call train API to retrain the movies suggestion model
	// Make GET request to suggestions API
	resp, err := http.Get(os.Getenv("MOVIES_ML_BASE_URL") + "/train")
	if err != nil {
		log.Println("failed to call train API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("train API returned status %d: %s", resp.StatusCode, string(body))
	}
}
