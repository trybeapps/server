package main

import (
    "fmt"
    "time"
    "database/sql"
    "net/http"
    "runtime"
    "math/rand"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "golang.org/x/crypto/bcrypt"
    _ "github.com/mattn/go-sqlite3"
    "gopkg.in/gomail.v2"
)

func main() {
    r := gin.Default()

    // Initiate session management (cookie-based)
    store := sessions.NewCookieStore([]byte("secret"))
    r.Use(sessions.Sessions("mysession", store))

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
    stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS `user` " +
        "(`id` INTEGER PRIMARY KEY AUTOINCREMENT, `name` VARCHAR(255) NOT NULL," +
        " `email` VARCHAR(255) UNIQUE NOT NULL, `password_hash` VARCHAR(255) NOT NULL," +
        " `confirmed` INTEGER DEFAULT 0)")
    CheckError(err)
    
    _, err = stmt.Exec()
    CheckError(err)

    // Create confirm table
    // Table: confirm
    // -----------------------------------------------------------------------------------------------------------
    // Fields: id, token, date_generated, date_expires, date_used, used, user_id (foreign key referencing user id)
    // -----------------------------------------------------------------------------------------------------------
    stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `confirm` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
        " `token` VARCHAR(255) NOT NULL, `date_generated` VARCHAR(255) NOT NULL," +
        " `date_expires` VARCHAR(255) NOT NULL, `date_used` VARCHAR(255), " +
        " `used` INTEGER DEFAULT 0, user_id INTEGER NOT NULL)")
    CheckError(err)

    _, err = stmt.Exec()
    CheckError(err)

    // Close sqlite3 database
    defer db.Close()

    // Router
    r.GET("/", GetHomePage)
    r.GET("/signin", GetSignIn)
    r.POST("/signin", PostSignIn)
    r.GET("/signup", GetSignUp)
    r.POST("/signup", PostSignUp)
    r.GET("/confirm-email", ConfirmEmail)
    r.GET("/new-token", SendNewToken)

    // Listen and serve on 0.0.0.0:8080
    r.Run(":8080")
}

func GetHomePage(c *gin.Context) {
    // Get session from cookie. Check if email exists
    // show Home page else redirect to signin page.
    session := sessions.Default(c)
    if session.Get("email") != nil {
        fmt.Println(session.Get("email"))
        c.HTML(200, "index.html", "")
    }
    c.Redirect(http.StatusMovedPermanently, "/signin")
}

func GetSignIn(c *gin.Context) {
    // Get session from cookie. Check if email exists
    // redirect to Home page else show signin page.
    session := sessions.Default(c)
    if session.Get("email") != nil {
        fmt.Println(session.Get("email"))
        c.Redirect(http.StatusMovedPermanently, "/")
    }
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

    db.Close()

    var hashedPassword []byte
    
    defer rows.Close()
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

        // Set cookie based session for signin
        session := sessions.Default(c)
        session.Set("email", email)
        session.Save()
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

    go SendEmail(int(id), name, email)

    c.String(http.StatusOK, "We have sent you a confirmation email for verification.")

}

// For confirm email token
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func SendEmail(id int, name string, email string) {

    // Set home many CPU cores this function wants to use.
    runtime.GOMAXPROCS(runtime.NumCPU())
    fmt.Println(runtime.NumCPU())

    token := randSeq(40)
    
    t := time.Now()
    
    dateGenerated := t.Format("20060102150405")
    fmt.Println("Token Date Generated: " + dateGenerated)
    
    dateExpires := t.AddDate(0,1,0).Format("20060102150405")
    fmt.Println("Token Date Expires: " + dateExpires)

    userId := id

    db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    stmt, err := db.Prepare("INSERT INTO confirm (token, date_generated, date_expires, user_id) VALUES (?, ?, ?, ?)")
    CheckError(err)

    _, err = stmt.Exec(token, dateGenerated, dateExpires, userId)
    CheckError(err)

    db.Close()

    confirmEmailLink := "http://localhost:8080/confirm-email?token=" + token

    m := gomail.NewMessage()
    m.SetHeader("From", "no-reply@libreread.org")
    m.SetHeader("To", email)
    // m.SetAddressHeader("Cc", "hello@nirm.al", "Nirmal")
    m.SetHeader("Subject", "LibreRead Email Confirmation")
    m.SetBody("text/html", "Hi " + name + 
        ",<br><br>Please confirm your email by clicking this link<br>" + 
        confirmEmailLink )
    // m.Attach("/home/Alex/lolcat.jpg")

    d := gomail.NewDialer("smtp.zoho.com", 587, "no-reply@libreread.org", "magicmode")

    // Send the confirmation email
    if err := d.DialAndSend(m); err != nil {
        panic(err)
    }
}

func ConfirmEmail(c * gin.Context) {
    token := c.Request.URL.Query()["token"][0]
    fmt.Println(token)

    db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    defer db.Close()

    // Get id from confirm table with the token got from url.
    rows, err := db.Query("select id, date_expires, user_id from confirm where token = ?", token)
    CheckError(err)

    var (
        id int
        dateExpires string
        userId int
    )
    
    if rows.Next() {
        err := rows.Scan(&id, &dateExpires, &userId)
        CheckError(err)

        fmt.Println(id)
        fmt.Println(dateExpires)
    } else {
        c.HTML(404, "invalid_token.html", "")
        return
    }
    rows.Close()

    t := time.Now()
    if currentDateTime := t.Format("20060102150405"); currentDateTime < dateExpires {
        stmt, err := db.Prepare("update confirm set date_used=?, used=? where id=?")
        CheckError(err)

        _, err = stmt.Exec(currentDateTime, 1, id)
        CheckError(err)

        stmt, err = db.Prepare("update user set confirmed=? where id=?")
        CheckError(err)

        _, err = stmt.Exec(1, userId)

        c.HTML(http.StatusOK, "confirmed.html", gin.H{
            "id": userId,
        })
        return
    } else {
        c.HTML(http.StatusOK, "expired.html", gin.H{
            "id": userId,
        })
        return
    }
}

func SendNewToken(c * gin.Context) {
    userId, err := strconv.Atoi(c.Request.URL.Query()["id"][0])
    CheckError(err)
    fmt.Println(userId)

    db, err := sql.Open("sqlite3", "./libreread.db")
    CheckError(err)

    defer db.Close()

    rows, err := db.Query("select name, email from user where id = ?", userId)
    CheckError(err)

    var (
        name string
        email string
    )
    
    if rows.Next() {
        err := rows.Scan(&name, &email)
        CheckError(err)

        fmt.Println(name)
        fmt.Println(email)
    }
    rows.Close()
    SendEmail(int(userId), name, email)
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}