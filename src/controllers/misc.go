package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func FireAuthConfig(c *gin.Context) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	firebaseKey := os.Getenv("ENVIRONMENT")
	if firebaseKey != "" && strings.ToLower(firebaseKey) == "prod" {
		firebaseKey = "prod"
	} else {
		firebaseKey = "dev"
	}

	configFilePath := filepath.Join(wd, "./../firebaseConfig_"+firebaseKey+".json")

	fileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error reading JSON file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read configuration file"})
		return
	}

	c.Data(http.StatusOK, "application/json", fileContent)
}
