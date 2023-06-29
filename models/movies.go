package models

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type Tabler interface {
	TableName() string
}

var ErrMovieNotOwned = errors.New("you can only delete your own movie")

// TableName overrides the table name used by User to `profiles`
func (Movie) TableName() string {
	return "watchlist"
}

type Movie struct {
	ID          uint    `gorm:"primary_key" json:"id"`
	UserID      uint    `json:"user_id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate *string `json:"release_date"`
	Image       string  `json:"image"`
	MovieID     uint    `json:"movie_id"`
	EmailSent   bool    `json:"email_sent"`
}

func GetMoviesByUserID(uid uint) ([]Movie, error) {
	var movies []Movie

	if err := DB.Find(&movies, "user_id = ?", uid).Error; err != nil {
		return movies, fmt.Errorf("watchlist for user id %d not found", uid)
	}

	return movies, nil

}

func (movie *Movie) UpdateMovie() error {
	return DB.Save(&movie).Error
}

func (movie *Movie) UpdateReleaseDate() {
	movieInfo, err := TMDbClient.GetMovieReleaseDates(int(movie.MovieID))

	if err != nil {
		log.Println("Error retrieving movie release dates", err)
	} else {
		var releaseDate *time.Time = nil
		for _, result := range movieInfo.MovieReleaseDatesResults.Results {
			for _, movieReleaseDate := range result.ReleaseDates {
				if movieReleaseDate.Type > 3 {
					relDate, err := time.Parse(time.RFC3339Nano, movieReleaseDate.ReleaseDate)
					if err != nil {
						log.Println("Error decoding date", movieReleaseDate.ReleaseDate)
					} else {
						if releaseDate == nil {
							releaseDate = &relDate
						} else {
							if releaseDate.After(relDate) {
								releaseDate = &relDate
							}
						}
					}
				}
			}
		}

		if releaseDate != nil {
			formattedDate := releaseDate.Format("2006-01-02")
			movie.ReleaseDate = &formattedDate
		}
	}
}

func GetMoviesWithoutReleaseDateByUserID(uid uint) []Movie {
	var wl []Movie

	DB.Where("release_date IS NULL").Find(&wl, "user_id = ?", uid)

	return wl
}

func (movie *Movie) SaveMovieToWatchlist() (*Movie, error) {
	movie.UpdateReleaseDate()
	if err := DB.Create(&movie).Error; err != nil {
		return &Movie{}, err
	}
	return movie, nil
}

func DeleteMovieFromWatchlistByID(id string, uid uint) error {
	var wl Movie

	if err := DB.First(&wl, id).Error; err != nil {
		return err
	}

	if wl.UserID != uid {
		return ErrMovieNotOwned
	}

	if err := DB.Delete(&wl).Error; err != nil {
		return err
	}

	return nil
}
