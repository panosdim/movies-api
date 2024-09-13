package controllers

import (
	"errors"
	"movies-backend/ai"
	"movies-backend/models"
	"movies-backend/utils"
	"movies-backend/utils/token"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetWatchlist(c *gin.Context) {
	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wl, err := models.GetWatchlistByUserID(userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wl)
}

func GetMovies(c *gin.Context) {
	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wl, err := models.GetMoviesByUserID(userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wl)
}

type WatchlistInput struct {
	Title    string `json:"title" binding:"required"`
	MovieId  uint   `json:"movie_id" binding:"required"`
	Image    string `json:"image"`
}

func AddToWatchlist(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input WatchlistInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wl := models.Movie{}

	wl.Title = input.Title
	wl.MovieID = input.MovieId
	wl.Image = input.Image
	wl.UserID = userId
	wl.ReleaseDate = nil

	newMovie, err := wl.SaveMovieToWatchlist()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.ClearUserMovieSuggestionCache(userId)

	c.JSON(http.StatusCreated, newMovie)
}

func DeleteFromWatchlist(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	if err := models.DeleteMovieFromWatchlistByID(id, userId); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrMovieNotOwned):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	utils.ClearUserMovieSuggestionCache(userId)

	c.JSON(http.StatusNoContent, nil)
}

func MarkMovieAsDownloaded(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	if err := models.MarkMovieAsDownloadedByID(id, userId); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrMovieNotOwned):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func MarkMovieAsWatched(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	if err := models.MarkMovieAsWatchedByID(id, userId); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrMovieNotOwned):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type RatingInput struct {
	Rating uint `json:"rating" binding:"required"`
}

func RateMovie(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")

	var input RatingInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.RateMovieByID(id, userId, input.Rating); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrMovieNotOwned):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	utils.ClearUserMovieSuggestionCache(userId)

	c.JSON(http.StatusNoContent, nil)
}

func MoviesSuggestion(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get numMovies from query parameter, default to 5 if not provided or invalid
	numMovies := 5
	if numMoviesStr := c.Query("numMovies"); numMoviesStr != "" {
		if _, err := strconv.Atoi(numMoviesStr); err == nil {
			numMovies, _ = strconv.Atoi(numMoviesStr)
		}
	}

	wl, err := ai.GetMoviesSuggestion(userId, numMovies)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wl)
}
