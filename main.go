package main

import (
	"fmt"
	"log"
	"movies-backend/controllers"
	"movies-backend/middlewares"
	"movies-backend/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	if err := r.Run(fmt.Sprintf(":%s", os.Getenv("PORT"))); err != nil {
		log.Fatalf("Error starting server")
	}
}
