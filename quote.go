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
        "quote": "So many things are possible just as long as you don't know they're impossible.",
        "author": "— Norton Juster",
        "image": "https://images.gr-assets.com/authors/1201117378p5/214.jpg",
    })
}