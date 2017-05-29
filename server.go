package main

import (
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// HTML rendering
	r.LoadHTMLGlob("templates/*")

	r.GET("/", GetHomePage)
	r.GET("/signin", GetSignIn)
	r.POST("/signin", PostSignIn)
	r.GET("/signup", GetSignUp)


	r.Run() // listen and serve on 0.0.0.0:8080
}

func GetHomePage(c *gin.Context) {
	c.HTML(200, "index.html", "")
}

func GetSignIn(c *gin.Context) {
	c.HTML(200, "signin.html", "")
}

func PostSignIn(c *gin.Context) {
	email := c.PostForm("email")
	password := []byte(c.PostForm("password"))

	fmt.Println(email)
	fmt.Println(password)

	// Hashing the password with the default cost of 10
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    CheckError(err)
    fmt.Println(string(hashedPassword))

    // Comparing the password with the hash
    err = bcrypt.CompareHashAndPassword(hashedPassword, password)
    fmt.Println(err) // nil means it is a match
}

func GetSignUp(c *gin.Context) {
	c.HTML(200, "signup.html", "")
}

func CheckError(err error) {
	if err != nil {
        panic(err)
    }
}