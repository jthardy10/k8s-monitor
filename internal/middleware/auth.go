package middleware

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

const API_KEY = "dev-api-key-123" // In production, this should be environment variable

func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
            c.Abort()
            return
        }

        if apiKey != API_KEY {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}
