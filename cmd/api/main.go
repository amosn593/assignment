package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/amosn593/assignment/docs"
	"github.com/amosn593/assignment/internal/router"
	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

// @title Clinic Booking System API
// @version 1.0
// @description REST API for clinic doctor appointments booking using Gin.
// @host localhost:8080
// @BasePath /
func main() {

	// Switch Gin to Release mode to suppress debug warnings
	gin.SetMode(gin.ReleaseMode)

	db, err := sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := router.SetupRouter(db)

	log.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}