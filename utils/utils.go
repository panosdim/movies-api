package utils

import (
	"fmt"
	"log"
	"movies-backend/models"
	"movies-backend/utils/mail"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
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

func GetTextResponse(resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates found in response")
	}

	// Initialize a strings.Builder to accumulate content from all candidates
	var allCandidatesContent strings.Builder

	// Process each candidate's content
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			// Handle different types of parts by type assertion
			switch p := part.(type) {
			case genai.Text:
				// Part is a Text type, accumulate the content
				allCandidatesContent.WriteString(string(p))
			default:
				return "", fmt.Errorf("unhandled part type: %T", part)
			}
		}
	}
	return allCandidatesContent.String(), nil
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
