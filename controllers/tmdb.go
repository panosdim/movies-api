package controllers

import (
	"movies-backend/models"
	"movies-backend/utils/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPopularMovies(c *gin.Context) {
	options := map[string]string{
		"language": "en-US",
		"page":     "1",
	}

	popularMovies, err := models.TMDbClient.GetMoviePopular(options)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, popularMovies)
}

func UpdateReleaseDates(c *gin.Context) {
	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find all movies without release release_date
	movies := models.GetMoviesWithoutReleaseDateByUserID(userId)

	for _, movie := range movies {
		movie.UpdateReleaseDate()
		_ = movie.UpdateMovie()
	}

	c.JSON(http.StatusNoContent, nil)
}

type SearchInput struct {
	Term string `form:"term" binding:"required"`
}

func SearchForMovie(c *gin.Context) {
	_, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input SearchInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	options := map[string]string{
		"language":      "en-US",
		"page":          "1",
		"include_adult": "false",
	}

	movies, _ := models.TMDbClient.GetSearchMovies(input.Term, options)

	c.JSON(http.StatusOK, movies)
}

func AutocompleteSearch(c *gin.Context) {
	_, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input SearchInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	options := map[string]string{
		"language":      "en-US",
		"page":          "1",
		"include_adult": "false",
	}

	movies, _ := models.TMDbClient.GetSearchMovies(input.Term, options)

	var autocompleteResults [][3]string
	const imageURL = "https://image.tmdb.org/t/p/w92"

	for _, movie := range movies.Results {
		if len(movie.PosterPath) > 0 {
			autocompleteResults = append(autocompleteResults, [3]string{movie.OriginalTitle, movie.ReleaseDate, imageURL + movie.PosterPath})
		} else {
			autocompleteResults = append(autocompleteResults, [3]string{movie.OriginalTitle, movie.ReleaseDate, movie.PosterPath})
		}
	}

	c.JSON(http.StatusOK, autocompleteResults)
}
