package main

import (
    "fmt"
    "time"
    "encoding/json"
    "net/http"
    "database/sql"
    "runtime"
    "math/rand"
    "strconv"
    "io"
    "mime"
    "os"
    "os/exec"
    "bytes"
    "strings"

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
    r.Static("/uploads", "./uploads")

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

    // Create book table
    // Table: book
    // ---------------------------------------------------------------------------------
    // Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on
    // ---------------------------------------------------------------------------------
    stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `book` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
        " `title` VARCHAR(255) NOT NULL, `filename` VARCHAR(255) NOT NULL," +
        " `author` VARCHAR(255) NOT NULL, `url` VARCHAR(255) NOT NULL," +
        " `cover` VARCHAR(255) NOT NULL, `pages` INTEGER NOT NULL, `current_page` INTEGER DEFAULT 0," +
        " `uploaded_on` VARCHAR(255) NOT NULL)")
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
    r.GET("/signout", GetSignOut)
    r.POST("/upload", PostUpload)
    r.GET("/collections", GetCollections)

    title, author, pages := GetPdfInfo()
    fmt.Println(title)
    fmt.Println(author)
    fmt.Println(pages)

    // GeneratePDFCover()

    // Listen and serve on 0.0.0.0:8080
    r.Run(":8080")
}

func GeneratePDFCover() {
    cmd := exec.Command("/usr/local/bin/pdfimages", "-p", "-png", "-f", "1", "-l", "2", 
        "uploads/Bird_Richard-Thinking_Functionally_with_Haskell-Cambridge_University_Press_2015_41371491567097.pdf", 
        "cover")

    err := cmd.Run()
    CheckError(err)
}

func GetPdfInfo() (string, string, string) {
    cmd := exec.Command("/usr/local/bin/pdfinfo", 
        "uploads/Bird_Richard-Thinking_Functionally_with_Haskell-Cambridge_University_Press_2015_41371491567097.pdf")
    
    var out bytes.Buffer
    cmd.Stdout = &out
    
    err :=  cmd.Run()
    CheckError(err)
    
    output := out.String()
    opSplit := strings.Split(output, "\n")
    
    title := opSplit[0]
    author := opSplit[1]
    pages := ""

    for _, element := range opSplit {
        if strings.HasPrefix(element, "Pages") {
            pages = strings.Split(element, ":")[1]
            pages = strings.Trim(pages, " ")
            break
        }
    }

    if strings.HasPrefix(title, "Title") {
        title = strings.Split(title, ":")[1]
        title = strings.Trim(title, " ")
    } else {
        title = ""
    }

    if strings.HasPrefix(author, "Author") {
        author = strings.Split(author, ":")[1]
        author = strings.Trim(author, " ")
    } else {
        author = ""
    }

    return title, author, pages
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}
var myClient = &http.Client{Timeout: 10 * time.Second}
func GetJSON(url string, target interface{}) error {
    r, _ := myClient.Get(url)
    if r != nil {
        defer r.Body.Close()
        return json.NewDecoder(r.Body).Decode(target)
    }
    return nil
}

type QS struct {
    Author string
    AuthorURL string
    FromBook string
    FromBookURL string
    Image string
    Quote string
}

func GetHomePage(c *gin.Context) {
    // Get session from cookie. Check if email exists
    // show Home page else redirect to signin page.
    session := sessions.Default(c)
    if session.Get("email") != nil {
        fmt.Println(session.Get("email"))

        q := new(QS)
        GetJSON("http://localhost:3000/quote-of-the-day", q)
        var quote, author, authorURL, image, fromBook, fromBookURL string   
        if q.Quote != "" {
            quote = q.Quote
            author = q.Author
            authorURL = q.AuthorURL
            image = q.Image
            fromBook = q.FromBook
            fromBookURL = q.FromBookURL
        } else {
            quote = "So many things are possible just as long as you don't know they're impossible."
            author = "Norton Juster"
            authorURL = "https://www.goodreads.com/author/show/214.Norton_Juster"
            image = "https://images.gr-assets.com/authors/1201117378p5/214.jpg"
            fromBook = "The Phantom Tollbooth"
            fromBookURL = "https://www.goodreads.com/work/1782584"
        }

        c.HTML(302, "index.html", gin.H{
            "qQuote": quote,
            "qAuthor": author,
            "qAuthorURL": authorURL,
            "qImage": image,
            "qFromBook": fromBook,
            "qFromBookURL": fromBookURL,
        })
    }
    c.Redirect(302, "/signin")
}

func GetSignIn(c *gin.Context) {
    // Get session from cookie. Check if email exists
    // redirect to Home page else show signin page.
    session := sessions.Default(c)
    if session.Get("email") != nil {
        fmt.Println(session.Get("email"))
        c.Redirect(302, "/")
    }
    c.HTML(302, "signin.html", "")
}

func GetSignOut(c *gin.Context) {
    session := sessions.Default(c)
    session.Delete("email")
    session.Save()

    c.Redirect(302, "/")
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
        c.Redirect(302, "/")

        // Set cookie based session for signin
        session := sessions.Default(c)
        session.Set("email", email)
        session.Save()
    } else {
        c.HTML(302, "signin.html", "")
    }
}

func GetSignUp(c *gin.Context) {
    c.HTML(302, "signup.html", "")
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

    go SendConfirmationEmail(int(id), name, email)

    c.HTML(302, "confirm_email.html", "")

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

func SendConfirmationEmail(id int, name string, email string) {

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

func ConfirmEmail(c *gin.Context) {
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

        c.HTML(302, "confirmed.html", gin.H{
            "id": userId,
        })
        return
    } else {
        c.HTML(302, "expired.html", gin.H{
            "id": userId,
        })
        return
    }
}

func SendNewToken(c *gin.Context) {
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
    go SendConfirmationEmail(int(userId), name, email)
}

func PostUpload(c *gin.Context) {
    multipart, err := c.Request.MultipartReader()
    CheckError(err)

    for {
        mimePart, err := multipart.NextPart()

        if err == io.EOF {
            break
        }

        CheckError(err)

        fmt.Println(mimePart)

        disposition, params, err := mime.ParseMediaType(mimePart.Header.Get("Content-Disposition"))
        CheckError(err)
        fmt.Println(disposition)
        fmt.Println(params["filename"])

        if contentType, _, _ := mime.ParseMediaType(mimePart.Header.Get("Content-Type")); contentType == "application/pdf" {
            out, err := os.Create("./uploads/" + params["filename"])
            CheckError(err)

            _, err = io.Copy(out, mimePart)
            CheckError(err)

            out.Close()

            // db, err := sql.Open("sqlite3", "./libreread.db")
            // CheckError(err)

            // ---------------------------------------------------------------------------------
            // Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on
            // ---------------------------------------------------------------------------------
            // stmt, err := db.Prepare("INSERT INTO book (title, filename, author, url, cover, pages, uploaded_on) VALUES (?, ?, ?)")
            // CheckError(err)

            // res, err := stmt.Exec(name, email, hashedPassword)
            // CheckError(err)

            // id, err := res.LastInsertId()
            // CheckError(err)

            // fmt.Println(id)

            // db.Close()
        }
    }

    c.String(200, "Books uploaded successfully")
}

func GetCollections(c *gin.Context) {
    c.HTML(302, "collections.html", "")
}