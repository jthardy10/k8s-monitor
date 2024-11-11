package main

import (
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
        })
    })
    
    r.GET("/clusters", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "clusters": []string{},
        })
    })
    
    log.Fatal(r.Run(":8080"))
}
