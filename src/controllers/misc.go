package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func FireAuthConfig(c *gin.Context) {
	firebaseConfig := map[string]string{
		"apiKey":            "AIzaSyCd7tCKU7uPd9iR9vmwl5UF8OfekcSbZyI",
		"authDomain":        "resttradingsystem.firebaseapp.com",
		"projectId":         "resttradingsystem",
		"storageBucket":     "resttradingsystem.appspot.com",
		"messagingSenderId": "635522974118",
		"appId":             "1:635522974118:web:0c045d0829049cc2a3cf67",
		"measurementId":     "G-10GXC9B3C8",
	}

	c.JSON(http.StatusOK, firebaseConfig)
}
