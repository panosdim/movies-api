package main

import (
	"fmt"
	"log"
	"movies-backend/controllers"
	"movies-backend/middlewares"
	"movies-backend/models"
	"movies-backend/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {

	models.ConnectDataBase()

	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())

	public := r.Group("/api")

	public.POST("/login", controllers.Login)

	public.GET("/popular", controllers.GetPopularMovies)

	public.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0"})
	})

	private := r.Group("/api")

	private.Use(middlewares.JwtAuthMiddleware())
	{
		private.GET("/user", controllers.CurrentUser)
		private.GET("/movies", controllers.GetWatchlist)
		private.POST("/movies", controllers.AddToWatchlist)
		private.DELETE("/movies/:id", controllers.DeleteFromWatchlist)
		private.GET("/update", controllers.UpdateReleaseDates)
		private.POST("/search", controllers.SearchForMovie)
		private.POST("/autocomplete", controllers.AutocompleteSearch)
	}

	// Schedule movies release date updates every day
	s := gocron.NewScheduler(time.UTC)
	if _, err := s.Every(60).Seconds().Do(func() { utils.CheckForAvailableMovies() }); err != nil {
		log.Fatalf("Error starting cron job")
	}
	s.StartAsync()

	if err := r.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		log.Fatalf("Error starting server")
	}
}
