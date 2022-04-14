package models

import (
	"fmt"
	"log"
	"os"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

var DB *gorm.DB
var TMDbClient *tmdb.Client

func ConnectDataBase() {
	var err error
	if err = godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	DbDriver := os.Getenv("DB_DRIVER")
	DbHost := os.Getenv("DB_HOST")
	DbUser := os.Getenv("DB_USER")
	DbPassword := os.Getenv("DB_PASSWORD")
	DbName := os.Getenv("DB_NAME")
	DbPort := os.Getenv("DB_PORT")

	DBUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=False&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	DB, err = gorm.Open(DbDriver, DBUrl)

	if err != nil {
		fmt.Println("Cannot connect to database ", DbDriver)
		log.Fatal("connection error:", err)
	} else {
		fmt.Println("We are connected to the database ", DbDriver)
	}

	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Movie{})

	// Initialize TMDb API library
	TMDbClient, err = tmdb.Init(os.Getenv("TMDB_KEY"))
	if err != nil {
		log.Fatal("TMDb connection error:", err)
	}
	TMDbClient.SetClientAutoRetry()
}
