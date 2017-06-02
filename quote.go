package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Serve static files
    r.Static("/static", "./static")

    // Router
    r.GET("/quote-of-the-day", GetQuote)

    // Listen and serve on 0.0.0.0:8080
    r.Run(":3000")
}

func GetQuote(c *gin.Context) {
    c.JSON(200, gin.H{
        "quote": "Don't cry because it's over, smile because it happened.",
        "author": "- Dr.Seuss",
        "image": "localhost:3000/static/img/seuss.jpg",
    })
}