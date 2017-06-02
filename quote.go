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
        "quote": "If you want to know what a man's like, take a good look at how he treats his inferiors, not his equals.",
        "author": "J.K. Rowling",
        "authorURL": "https://www.goodreads.com/author/show/1077326.J_K_Rowling",
        "image": "https://images.gr-assets.com/authors/1415945171p2/1077326.jpg",
        "fromBook": "Harry Potter and the Goblet of Fire",
        "fromBookURL": "https://www.goodreads.com/book/show/6.Harry_Potter_and_the_Goblet_of_Fire",
    })
}