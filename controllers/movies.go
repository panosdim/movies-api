package controllers

import (
	"errors"
	"movies-backend/models"
	"movies-backend/utils/token"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetWatchlist(c *gin.Context) {
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
	Overview string `json:"overview" binding:"required"`
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
	wl.Overview = input.Overview
	wl.MovieID = input.MovieId
	wl.Image = input.Image
	wl.UserID = userId
	wl.ReleaseDate = nil

	newMovie, err := wl.SaveMovieToWatchlist()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	c.JSON(http.StatusNoContent, nil)
}
