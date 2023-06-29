package utils

import (
	"log"
	"movies-backend/models"
	"movies-backend/utils/mail"
	"time"
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
