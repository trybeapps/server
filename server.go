package main

import (
	"fmt"
	"database/sql"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// HTML rendering
	r.LoadHTMLGlob("templates/*")

	// Open sqlite3 database
	db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    // Create user table
    // Table: user
    // -------------------------------------------------
    // Fields: id, name, email, password_hash, confirmed
    // -------------------------------------------------
    stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS `user` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `name` VARCHAR(255) NOT NULL, `email` VARCHAR(255) UNIQUE NOT NULL, `password_hash` VARCHAR(255) NOT NULL, `confirmed` INTEGER DEFAULT 0)")
    CheckError(err)
    
    _, err = stmt.Exec()
    CheckError(err)

    // Close sqlite3 database
    db.Close()

    // Router
	r.GET("/", GetHomePage)
	r.GET("/signin", GetSignIn)
	r.POST("/signin", PostSignIn)
	r.GET("/signup", GetSignUp)
	r.POST("/signup", PostSignUp)

	// Listen and serve on 0.0.0.0:8080
	r.Run()
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

	db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    rows, err := db.Query("select password_hash from user where email = ?", email)
    CheckError(err)

    var hashedPassword []byte
    
    if rows.Next() {
		err := rows.Scan(&hashedPassword)
		CheckError(err)
		fmt.Println(hashedPassword)
	}

    // Comparing the password with the hash
    err = bcrypt.CompareHashAndPassword(hashedPassword, password)
    fmt.Println(err) // nil means it is a match

    if err == nil {
    	c.Redirect(http.StatusMovedPermanently, "/")
    } else {
    	c.HTML(200, "signin.html", "")
    }
}

func GetSignUp(c *gin.Context) {
	c.HTML(200, "signup.html", "")
}

func PostSignUp(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := []byte(c.PostForm("password"))

	fmt.Println(name)
	fmt.Println(email)
	fmt.Println(password)

	// Hashing the password with the default cost of 10
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    CheckError(err)
    fmt.Println(string(hashedPassword))

	db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    stmt, err := db.Prepare("INSERT INTO user (name, email, password_hash) VALUES (?, ?, ?)")
    CheckError(err)

    res, err := stmt.Exec(name, email, hashedPassword)
    CheckError(err)

    id, err := res.LastInsertId()
    CheckError(err)

    fmt.Println(id)

    db.Close()
}

func CheckError(err error) {
	if err != nil {
        panic(err)
    }
}