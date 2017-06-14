package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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
		" `used` INTEGER DEFAULT 0, `user_id` INTEGER NOT NULL)")
	CheckError(err)

	_, err = stmt.Exec()
	CheckError(err)

	// Create book table
	// Table: book
	// ------------------------------------------------------------------------------------------
	// Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on, user_id
	// ------------------------------------------------------------------------------------------
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `book` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		" `title` VARCHAR(255) NOT NULL, `filename` VARCHAR(255) NOT NULL," +
		" `author` VARCHAR(255) NOT NULL, `url` VARCHAR(255) NOT NULL," +
		" `cover` VARCHAR(255) NOT NULL, `pages` INTEGER NOT NULL, `current_page` INTEGER DEFAULT 0," +
		" `uploaded_on` VARCHAR(255) NOT NULL, `user_id` INTEGER NOT NULL)")
	CheckError(err)

	_, err = stmt.Exec()
	CheckError(err)

	// Create currently_reading table
	// Table: currently_reading
	// -------------------
	// Fields: id, user_id
	// -------------------
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `currently_reading` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		" `book_id` INTEGER NOT NULL, `user_id` INTEGER NOT NULL, `date_read` VARCHAR(255) NOT NULL)")
	CheckError(err)

	_, err = stmt.Exec()
	CheckError(err)

	// Create collection table
	// Table: collection
	// ----------------------------------------------
	// Fields: id, title, description, books, user_id
	// ----------------------------------------------
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `collection` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		" `title` VARCHAR(255) NOT NULL, `description` VARCHAR(1200) NOT NULL, `books` VARCHAR(1200) NOT NULL," +
		" `cover` VARCHAR(255) NULL, `user_id` INTEGER NOT NULL)")
	CheckError(err)

	_, err = stmt.Exec()
	CheckError(err)

	// Close sqlite3 database
	db.Close()

	// Init Elasticsearch attachment
	attachment := &AS{
		Description: "Process documents",
		Processors: []ASP{
			ASP{
				Attachment: ASPA{
					Field:        "thedata",
					IndexedChars: -1,
				},
			},
		},
	}

	fmt.Println(attachment)

	b, err := json.Marshal(attachment)
	CheckError(err)
	fmt.Println(b)

	PutJSON("http://localhost:9200/_ingest/pipeline/attachment", b)

	// Init Elasticsearch index
	index := &IS{
		ISS{
			NumberOfShards:   4,
			NumberOfReplicas: 0,
		},
	}

	b, err = json.Marshal(index)
	CheckError(err)
	fmt.Println(b)

	PutJSON("http://localhost:9200/lr_index", b)

	// Router
	r.GET("/", GetHomePage)
	r.GET("/signin", GetSignIn)
	r.POST("/signin", PostSignIn)
	r.GET("/signup", GetSignUp)
	r.POST("/signup", PostSignUp)
	r.GET("/confirm-email", ConfirmEmail)
	r.GET("/new-token", SendNewToken)
	r.GET("/signout", GetSignOut)
	r.POST("/upload", UploadBook)
	r.GET("/book/:bookname", SendBook)
	r.GET("/cover/:covername", SendBookCover)
	r.GET("/books/:pagination", GetPagination)
	r.GET("/autocomplete", GetAutocomplete)
	r.GET("/collections", GetCollections)
	r.GET("/add-collection", GetAddCollection)
	r.POST("/post-new-collection", PostNewCollection)
	r.GET("/collection/:id", GetCollection)

	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")
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

func PutJSON(url string, message []byte) {
	fmt.Println(url)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(message))
	CheckError(err)
	res, err := myClient.Do(req)
	CheckError(err)
	content, err := ioutil.ReadAll(res.Body)
	CheckError(err)
	fmt.Println(string(content))
}

type AS struct {
	Description string `json:"description"`
	Processors  []ASP  `json:"processors"`
}

type ASP struct {
	Attachment ASPA `json:"attachment"`
}

type ASPA struct {
	Field        string `json:"field"`
	IndexedChars int64  `json:"indexed_chars"`
}

type IS struct {
	Settings ISS `json:"settings"`
}

type ISS struct {
	NumberOfShards   int64 `json:"number_of_shards"`
	NumberOfReplicas int64 `json:"number_of_replicas"`
}

func SendBook(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		name := c.Param("bookname")
		fmt.Println(name)

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("SELECT `id` FROM `user` WHERE `email` = ?", session.Get("email"))
		CheckError(err)

		var userId int64
		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)
		}
		fmt.Println(userId)
		rows.Close()

		rows, err = db.Query("SELECT `id` FROM `book` WHERE `filename` = ?", name)
		CheckError(err)

		var bookId int64
		if rows.Next() {
			err := rows.Scan(&bookId)
			CheckError(err)
		}
		fmt.Println(bookId)
		rows.Close()

		t := time.Now()

		dateRead := t.Format("20060102150405")
		fmt.Println("Date read: " + dateRead)

		rows, err = db.Query("SELECT `id` FROM `currently_reading` WHERE `book_id` = ?", bookId)
		CheckError(err)

		var currentlyReadingId int64
		if rows.Next() {
			err := rows.Scan(&currentlyReadingId)
			CheckError(err)
		}
		fmt.Println(currentlyReadingId)
		rows.Close()

		if currentlyReadingId == 0 {
			stmt, err := db.Prepare("INSERT INTO `currently_reading` (book_id, user_id, date_read) VALUES (?, ?, ?)")
			CheckError(err)

			res, err := stmt.Exec(bookId, userId, dateRead)
			CheckError(err)

			id, err := res.LastInsertId()
			CheckError(err)

			fmt.Println(id)
		} else {
			stmt, err := db.Prepare("UPDATE `currently_reading` SET date_read=? WHERE id=?")
			CheckError(err)

			_, err = stmt.Exec(dateRead, currentlyReadingId)
			CheckError(err)
		}

		c.HTML(200, "viewer.html", "")
	}
	c.Redirect(302, "/signin")
}

func SendBookCover(c *gin.Context) {
	name := c.Param("covername")
	filePath := "./uploads/img/" + name

	c.File(filePath)
}

type QS struct {
	Author      string
	AuthorURL   string
	FromBook    string
	FromBookURL string
	Image       string
	Quote       string
}

type BS struct {
	Title string
	URL   string
	Cover string
}

type BSList []BS

func GetHomePage(c *gin.Context) {
	// Get session from cookie. Check if email exists
	// show Home page else redirect to signin page.
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		q := new(QS)
		GetJSON("http://localhost:3000/quote-of-the-day", q)

		if q.Quote == "" {
			q.Quote = "So many things are possible just as long as you don't know they're impossible."
			q.Author = "Norton Juster"
			q.AuthorURL = "https://www.goodreads.com/author/show/214.Norton_Juster"
			q.Image = "https://images.gr-assets.com/authors/1201117378p5/214.jpg"
			q.FromBook = "The Phantom Tollbooth"
			q.FromBookURL = "https://www.goodreads.com/work/1782584"
		}

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("SELECT `id` FROM `user` WHERE `email` = ?", session.Get("email"))
		CheckError(err)

		var id int64
		if rows.Next() {
			err := rows.Scan(&id)
			CheckError(err)
		}
		fmt.Println(id)
		rows.Close()

		// Get currently reading books.
		rows, err = db.Query("SELECT `book_id` FROM `currently_reading` WHERE `user_id` = ? ORDER BY `date_read` DESC LIMIT ?, ?", id, 0, 12)
		CheckError(err)

		var crBooks []int64
		for rows.Next() {
			var crBook int64
			err = rows.Scan(&crBook)
			CheckError(err)

			crBooks = append(crBooks, crBook)
		}
		fmt.Println(crBooks)
		rows.Close()

		// Get book title, url, cover for currently reading books.
		crb := []BS{}
		for _, num := range crBooks {
			rows, err = db.Query("SELECT `title`, `url`, `cover` FROM `book` WHERE `id` = ?", num)
			CheckError(err)

			var (
				title string
				url   string
				cover string
			)
			if rows.Next() {
				err = rows.Scan(
					&title,
					&url,
					&cover,
				)
				CheckError(err)

				crb = append(crb, BS{
					title,
					url,
					cover,
				})
			}
			rows.Close()
		}
		fmt.Println(crb)

		// Check total number of rows in book table
		rows, err = db.Query("SELECT COUNT(*) AS count FROM `book` WHERE `user_id` = ?", id)
		CheckError(err)

		var count int64
		for rows.Next() {
			err = rows.Scan(&count)
			CheckError(err)
		}
		fmt.Println(count)

		var totalPages float64 = float64(float64(count) / 18.0)
		totalPagesDecimal := fmt.Sprintf("%.1f", totalPages)

		var tp int64
		if strings.Split(totalPagesDecimal, ".")[1] == "0" {
			tp = int64(totalPages)
		} else {
			tp = int64(totalPages) + 1
		}
		fmt.Println(tp)

		// ------------------------------------------------------------------------------------------
		// Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on, user_id
		// ------------------------------------------------------------------------------------------
		rows, err = db.Query("SELECT `title`, `url`, `cover` FROM `book` WHERE `user_id` = ? ORDER BY `id` DESC LIMIT ?, ?", id, 0, 18)
		CheckError(err)

		b := []BS{}

		var (
			title string
			url   string
			cover string
		)
		for rows.Next() {
			err = rows.Scan(
				&title,
				&url,
				&cover,
			)
			CheckError(err)

			b = append(b, BS{
				title,
				url,
				cover,
			})
		}

		rows.Close()
		db.Close()

		booksList := []BSList{}
		for i := 0; i < len(b); i += 6 {
			j := i + 6
			for j > len(b) {
				j -= 1
			}
			booksList = append(booksList, b[i:j])
		}

		booksListMedium := []BSList{}
		for i := 0; i < len(b); i += 3 {
			j := i + 3
			for j > len(b) {
				j -= 1
			}
			booksListMedium = append(booksListMedium, b[i:j])
		}

		booksListSmall := []BSList{}
		for i := 0; i < len(b); i += 2 {
			j := i + 2
			for j > len(b) {
				j -= 1
			}
			booksListSmall = append(booksListSmall, b[i:j])
		}

		booksListXtraSmall := b

		c.HTML(302, "index.html", gin.H{
			"q":                  q,
			"crb":                crb,
			"booksList":          booksList,
			"booksListMedium":    booksListMedium,
			"booksListSmall":     booksListSmall,
			"booksListXtraSmall": booksListXtraSmall,
			"tp":                 tp,
		})
	}
	c.Redirect(302, "/signin")
}

func GetPagination(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))
		pagination, err := strconv.Atoi(c.Param("pagination"))
		CheckError(err)

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("SELECT `id` FROM `user` WHERE `email` = ?", session.Get("email"))
		CheckError(err)

		var id int64
		if rows.Next() {
			err := rows.Scan(&id)
			CheckError(err)
		}
		fmt.Println(id)
		rows.Close()

		// Check total number of rows in book table
		rows, err = db.Query("SELECT COUNT(*) AS count FROM `book`")
		CheckError(err)

		var count int64
		for rows.Next() {
			err = rows.Scan(&count)
			CheckError(err)
		}
		fmt.Println(count)

		var totalPages float64 = float64(float64(count) / 18.0)
		totalPagesDecimal := fmt.Sprintf("%.1f", totalPages)

		var tp int64
		if strings.Split(totalPagesDecimal, ".")[1] == "0" {
			tp = int64(totalPages)
		} else {
			tp = int64(totalPages) + 1
		}
		fmt.Println(tp)

		// ------------------------------------------------------------------------------------------
		// Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on, user_id
		// ------------------------------------------------------------------------------------------
		rows, err = db.Query("SELECT `title`, `url`, `cover` FROM `book` WHERE `user_id` = ? ORDER BY `id` DESC LIMIT ?, ?", id, (pagination-1)*18, 18)
		CheckError(err)

		b := []BS{}

		var (
			title string
			url   string
			cover string
		)
		for rows.Next() {
			err = rows.Scan(
				&title,
				&url,
				&cover,
			)
			CheckError(err)

			b = append(b, BS{
				title,
				url,
				cover,
			})
		}
		rows.Close()
		db.Close()

		booksList := []BSList{}
		for i := 0; i < len(b); i += 6 {
			j := i + 6
			for j > len(b) {
				j -= 1
			}
			booksList = append(booksList, b[i:j])
		}

		booksListMedium := []BSList{}
		for i := 0; i < len(b); i += 3 {
			j := i + 3
			for j > len(b) {
				j -= 1
			}
			booksListMedium = append(booksListMedium, b[i:j])
		}

		booksListSmall := []BSList{}
		for i := 0; i < len(b); i += 2 {
			j := i + 2
			for j > len(b) {
				j -= 1
			}
			booksListSmall = append(booksListSmall, b[i:j])
		}

		booksListXtraSmall := b

		c.HTML(302, "pagination.html", gin.H{
			"pagination":         pagination,
			"booksList":          booksList,
			"booksListMedium":    booksListMedium,
			"booksListSmall":     booksListSmall,
			"booksListXtraSmall": booksListXtraSmall,
			"tp":                 tp,
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

	go SendConfirmationEmail(int64(id), name, email)

	c.HTML(302, "confirm_email.html", "")

}

// For confirm email token
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int64) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SendConfirmationEmail(id int64, name string, email string) {

	// Set home many CPU cores this function wants to use.
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())

	token := randSeq(40)

	t := time.Now()

	dateGenerated := t.Format("20060102150405")
	fmt.Println("Token Date Generated: " + dateGenerated)

	dateExpires := t.AddDate(0, 1, 0).Format("20060102150405")
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
	m.SetBody("text/html", "Hi "+name+
		",<br><br>Please confirm your email by clicking this link<br>"+
		confirmEmailLink)
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.zoho.com", 587, "no-reply@libreread.org", "magicmode")

	// Send the confirmation email
	if err := d.DialAndSend(m); err != nil {
		log.Fatal(err)
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
		id          int64
		dateExpires string
		userId      int64
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
		CheckError(err)

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
		name  string
		email string
	)

	if rows.Next() {
		err := rows.Scan(&name, &email)
		CheckError(err)

		fmt.Println(name)
		fmt.Println(email)
	}
	rows.Close()
	go SendConfirmationEmail(int64(userId), name, email)
}

func UploadBook(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("select id from user where email = ?", session.Get("email"))
		CheckError(err)

		var userId int64

		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)

			userIdString := fmt.Sprintf("%v", userId)
			fmt.Println("User id: " + userIdString)
		}
		rows.Close()

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
			contentType, _, _ := mime.ParseMediaType(mimePart.Header.Get("Content-Type"))

			// Store file in /uploads/ dir
			var fileName string
			if contentType == "application/pdf" {
				fileName = strings.Split(params["filename"], ".pdf")[0]
				fileName = strings.Join(strings.Split(fileName, " "), "_") + ".pdf"
			} else if contentType == "application/epub+zip" {
				fileName = strings.Split(params["filename"], ".epub")[0]
				fileName = strings.Join(strings.Split(fileName, " "), "_") + ".epub"
			}
			fmt.Println("filename: " + fileName)

			rows, err = db.Query("select id from book where filename = ?", fileName)
			CheckError(err)

			var bookId int64

			if rows.Next() {
				err := rows.Scan(&bookId)
				CheckError(err)

				bookIdString := fmt.Sprintf("%v", bookId)
				fmt.Println("Book id: " + bookIdString)
			}
			rows.Close()

			if bookId != 0 {
				c.String(200, fileName+" already exists. ")
				continue
			}

			if contentType == "application/pdf" {
				filePath := "./uploads/" + fileName

				out, err := os.Create(filePath)
				CheckError(err)

				_, err = io.Copy(out, mimePart)
				CheckError(err)

				out.Close()

				title, author, pages := GetPDFInfo(filePath)

				if title == "" {
					title = fileName
				}

				if author == "" {
					author = "unknown"
				}

				pagesInt, err := strconv.ParseInt(pages, 10, 64)
				CheckError(err)

				fmt.Println("Book title: " + title)
				fmt.Println("Book author: " + author)
				fmt.Println("Total pages: " + pages)

				url := "/book/" + fileName
				fmt.Println("Book URL: " + url)

				coverPath := "./uploads/img/" + fileName

				GeneratePDFCover(filePath, coverPath)

				cover := ""

				if _, err := os.Stat(coverPath + "-001-000.png"); err == nil {
					cover = "/cover/" + fileName + "-001-000.png"
				}

				fmt.Println("Book cover URL: " + cover)

				t := time.Now()

				uploadedOn := t.Format("20060102150405")
				fmt.Println("Uploaded on: " + uploadedOn)

				// ------------------------------------------------------------------------------------------
				// Fields: id, title, filename, author, url, cover, pages, current_page, uploaded_on, user_id
				// ------------------------------------------------------------------------------------------
				stmt, err := db.Prepare("INSERT INTO book (title, filename, author, url, cover, pages, uploaded_on, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
				CheckError(err)

				res, err := stmt.Exec(title, fileName, author, url, cover, pagesInt, uploadedOn, userId)
				CheckError(err)

				id, err := res.LastInsertId()
				CheckError(err)

				fmt.Println(id)

				// Feed book info to ES
				bookInfo := BIS{
					Title:  title,
					Author: author,
					URL:    url,
					Cover:  cover,
				}

				fmt.Println(bookInfo)

				indexURL := "http://localhost:9200/lr_index/book_info/" + strconv.Itoa(int(id))
				fmt.Println(indexURL)

				b, err := json.Marshal(bookInfo)
				CheckError(err)

				PutJSON(indexURL, b)

				// Feed book content to ES
				go FeedContent(filePath, userId, id, title, author, url, cover, pagesInt)

				c.String(200, fileName+" uploaded successfully. ")
			} else if contentType == "application/epub+zip" {
				filePath := "./uploads/" + fileName

				out, err := os.Create(filePath)
				CheckError(err)

				_, err = io.Copy(out, mimePart)
				CheckError(err)

				out.Close()

				fileName = strings.Split(fileName, ".epub")[0]
				go EPUBUnzip(filePath, fileName)

				fileUnzipPath := "./uploads/" + fileName + "/META-INF/container.xml"
				XMLContent, err := ioutil.ReadFile(fileUnzipPath)
				CheckError(err)

				var xxx = `
                <?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="en" lang="en">
    <head>
        <title>Chapter 1. Introduction</title>
        <link rel="stylesheet" type="text/css" href="css/epub.css" />
    </head>
    <body>
        <section class="chapter" title="Chapter 1. Introduction" epub:type="chapter" id="introduction">
            <h2 class="title">Chapter 1. Introduction</h2>
            <p>If you’re expecting a run-of-the-mill best practices manual, be aware that there’s an
                ulterior message that will be running through this one. While the primary goal is
                certainly to give you the information you need to create accessible EPUB 3
                publications, it also seeks to address the question of why you need to pay attention
                to the quality of your data, and how accessible data and general good data practices
                are more tightly entwined than you might think.</p>
        </section>
    </body>
</html>
`

				v := XMLCS{}
				err = xml.Unmarshal(XMLContent, &v)
				CheckError(err)
				fmt.Println(v)

				vv := CXMLS{}
				err = xml.Unmarshal([]byte(xxx), &vv)
				CheckError(err)
				fmt.Println(vv)
			}
		}
		db.Close()
	}
}

type CXMLS struct {
	Body CBXMLS `xml:"body"`
}

type CBXMLS struct {
	Content string `xml:",innerxml"`
}

type XMLCS struct {
	RootFiles XMLRFS `xml:"rootfiles"`
}

type XMLRFS struct {
	RootFile XMLRF `xml:"rootfile"`
}

type XMLRF struct {
	FullPath string `xml:"full-path,attr"`
}

type BIS struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	URL    string `json:"url"`
	Cover  string `json:"cover"`
}

func FeedContent(filePath string, userId int64, bookId int64, title string, author string, url string, cover string, pagesInt int64) {
	// Set home many CPU cores this function wants to use.
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())

	t := time.Now()
	timeNow := t.Format("20060102150405")
	path := "./uploads/splitpdf_" + strconv.Itoa(int(userId)) + "_" + timeNow
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}
	fmt.Println(path)
	fmt.Println(filePath)

	fmt.Println("\n\n\n\n\n")
	fmt.Println(title)
	var wg sync.WaitGroup
	wg.Add(1)
	go PDFSeparate(path, filePath, &wg)
	wg.Wait()
	fmt.Println("wg done!")

	var i int64
	for i = 1; i < (pagesInt + 1); i += 1 {
		pagePath := path + "/" + strconv.Itoa(int(i)) + ".pdf"
		if _, err := os.Stat(pagePath); os.IsNotExist(err) {
			continue
		}
		data, err := ioutil.ReadFile(pagePath)
		CheckError(err)

		sEnc := base64.StdEncoding.EncodeToString([]byte(string(data)))

		bookDetail := BDS{
			TheData: sEnc,
			Title:   title,
			Author:  author,
			URL:     url,
			Cover:   cover,
			Page:    i,
		}

		b, err := json.Marshal(bookDetail)
		CheckError(err)

		indexURL := "http://localhost:9200/lr_index/book_detail/" +
			strconv.Itoa(int(userId)) + "_" + strconv.Itoa(int(bookId)) +
			"_" + strconv.Itoa(int(i)) + "?pipeline=attachment"
		fmt.Println("Index URL: " + indexURL)
		PutJSON(indexURL, b)
	}

	// Remove the splitted files as it is not needed now.
	defer os.RemoveAll(path)
}

type BDS struct {
	TheData string `json:"thedata"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	URL     string `json:"url"`
	Cover   string `json:"cover"`
	Page    int64  `json:"page"`
}

func PDFSeparate(path string, filePath string, wg *sync.WaitGroup) error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())
	cmd := exec.Command("pdfseparate", filePath, path+"/%d.pdf")

	err := cmd.Start()
	CheckError(err)
	fmt.Println("Waiting for command to finish...")
	err = cmd.Wait()
	fmt.Printf("Command finished with error: %v", err)
	wg.Done()
	return nil
}

func EPUBUnzip(filePath string, fileName string) error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())
	cmd := exec.Command("unzip", filePath, "-d", "uploads/"+fileName+"/")

	err := cmd.Start()
	CheckError(err)
	fmt.Println("Waiting for command to finish...")
	err = cmd.Wait()
	fmt.Printf("Command finished with error: %v", err)
	return nil
}

func GeneratePDFCover(filePath, coverPath string) {
	cmd := exec.Command("pdfimages", "-p", "-png", "-f", "1", "-l", "2", filePath, coverPath)

	err := cmd.Run()
	CheckError(err)
}

func GetPDFInfo(filePath string) (string, string, string) {
	cmd := exec.Command("pdfinfo", filePath)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	CheckError(err)

	output := out.String()
	opSplit := strings.Split(output, "\n")

	title := opSplit[0]
	author := opSplit[1]
	pages := ""

	// Get number of pages.
	for _, element := range opSplit {
		if strings.HasPrefix(element, "Pages") {
			pages = strings.Split(element, ":")[1]
			pages = strings.Trim(pages, " ")
			break
		}
	}

	// Get book title.
	if strings.HasPrefix(title, "Title") {
		title = strings.Split(title, ":")[1]
		title = strings.Trim(title, " ")
	} else {
		title = ""
	}

	// Get author of the uploaded book.
	if strings.HasPrefix(author, "Author") {
		author = strings.Split(author, ":")[1]
		author = strings.Trim(author, " ")
	} else {
		author = ""
	}

	return title, author, pages
}

type BIP struct {
	Source []string `json:"_source"`
	Query  BIPQ     `json:"query"`
}

type BIPQ struct {
	MultiMatch MMQ `json:"multi_match"`
}

type MMQ struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

type BDP struct {
	Source    []string `json:"_source"`
	Query     BDPQ     `json:"query"`
	Highlight BDPH     `json:"highlight"`
}

type BDPQ struct {
	MatchPhrase BDPQAC `json:"match_phrase"`
}

type BDPQAC struct {
	AttachmentContent string `json:"attachment.content"`
}

type BDPH struct {
	Fields BDPHF `json:"fields"`
}

type BDPHF struct {
	AttachmentContent BDPHFAC `json:"attachment.content"`
}

type BDPHFAC struct {
	FragmentSize      int64 `json:"fragment_size"`
	NumberOfFragments int64 `json:"number_of_fragments"`
	NoMatchSize       int64 `json:"no_match_size"`
}

func GetAutocomplete(c *gin.Context) {
	q := c.Request.URL.Query()
	term := q["term"][0]
	fmt.Println(term)

	payloadInfo := &BIP{
		Source: []string{"title", "author", "url", "cover"},
		Query: BIPQ{
			MultiMatch: MMQ{
				Query:  term,
				Fields: []string{"title", "author"},
			},
		},
	}

	b, err := json.Marshal(payloadInfo)
	CheckError(err)

	indexURL := "http://localhost:9200/lr_index/book_info/_search"
	fmt.Println("Index URL: " + indexURL)

	res := GetJSONPassPayload(indexURL, b)
	target := BIRS{}
	json.Unmarshal(res, &target)

	hits := target.Hits.Hits
	hitsBIS := []BIS{}
	for _, el := range hits {
		hitsBIS = append(hitsBIS, BIS{
			Title:  el.Source.Title,
			Author: el.Source.Author,
			URL:    el.Source.URL,
			Cover:  el.Source.Cover,
		})
	}

	payloadDetail := &BDP{
		Source: []string{"title", "author", "url", "cover", "page"},
		Query: BDPQ{
			MatchPhrase: BDPQAC{
				AttachmentContent: term,
			},
		},
		Highlight: BDPH{
			Fields: BDPHF{
				AttachmentContent: BDPHFAC{
					FragmentSize:      150,
					NumberOfFragments: 3,
					NoMatchSize:       150,
				},
			},
		},
	}
	b, err = json.Marshal(payloadDetail)
	CheckError(err)

	indexURL = "http://localhost:9200/lr_index/book_detail/_search"
	fmt.Println("Index URL: " + indexURL)

	res = GetJSONPassPayload(indexURL, b)
	target2 := BDRS{}
	json.Unmarshal(res, &target2)

	hits2 := target2.Hits.Hits
	hitsBDRS := []BDRSHS{}
	for _, el := range hits2 {
		fmt.Println(el.Source)
		fmt.Println(el.Highlight)
		hitsBDRS = append(hitsBDRS, BDRSHS{
			Source:    el.Source,
			Highlight: el.Highlight,
		})
	}
	fmt.Println(hitsBDRS)

	bsr := BSR{
		BookInfo:   hitsBIS,
		BookDetail: hitsBDRS,
	}

	c.JSON(200, bsr)
}

type BSR struct {
	BookInfo   []BIS    `json:"book_info"`
	BookDetail []BDRSHS `json:"book_detail"`
}

type BDRS struct {
	Hits BDRSH `json:"hits"`
}

type BDRSH struct {
	Hits []BDRSHS `json:"hits"`
}

type BDRSHS struct {
	Source    BDS2 `json:"_source"`
	Highlight BDH  `json:"highlight"`
}

type BDH struct {
	AttachmentContent []string `json:"attachment.content"`
}

type BDS2 struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	URL    string `json:"url"`
	Cover  string `json:"cover"`
	Page   int64  `json:"page"`
}

type BIRS struct {
	Hits BIRSH `json:"hits"`
}

type BIRSH struct {
	Hits []BIRSHH `json:"hits"`
}

type BIRSHH struct {
	Source BIS `json:"_source"`
}

func GetJSONPassPayload(url string, payload []byte) []byte {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	CheckError(err)
	res, err := myClient.Do(req)
	CheckError(err)
	content, err := ioutil.ReadAll(res.Body)
	CheckError(err)
	fmt.Println(string(content))
	return content
}

func GetCollections(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("select id from user where email = ?", session.Get("email"))
		CheckError(err)

		var userId int64

		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)

			userIdString := fmt.Sprintf("%v", userId)
			fmt.Println("User id: " + userIdString)
		}
		rows.Close()

		rows, err = db.Query("select id, title, description, books, cover from collection where user_id = ?", userId)
		CheckError(err)

		cbks := []CBKS{}
		for rows.Next() {
			var (
				id          int64
				title       string
				description string
				books       string
				cover       sql.NullString
			)
			err := rows.Scan(&id, &title, &description, &books, &cover)
			CheckError(err)

			var c string
			if cover.Valid {
				c = cover.String
			} else {
				c = ""
			}

			cbks = append(cbks, CBKS{
				Id:          id,
				Title:       title,
				Description: description,
				Books:       books,
				Cover:       c,
			})
		}
		fmt.Println(cbks)

		db.Close()
		c.HTML(302, "collections.html", gin.H{
			"cbks": cbks,
		})
	}
	c.Redirect(302, "/signin")
}

type CBKS struct {
	Id          int64
	Title       string
	Description string
	Books       string
	Cover       string
}

func GetAddCollection(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("select id from user where email = ?", session.Get("email"))
		CheckError(err)

		var userId int64

		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)

			userIdString := fmt.Sprintf("%v", userId)
			fmt.Println("User id: " + userIdString)
		}
		rows.Close()

		rows, err = db.Query("select id, cover from book where user_id = ?", userId)
		CheckError(err)

		books := []BCSL{}

		for rows.Next() {
			var (
				b int64
				c string
			)
			err := rows.Scan(&b, &c)
			CheckError(err)

			books = append(books, BCSL{
				BookId: b,
				Cover:  c,
			})
		}
		fmt.Println(books)

		db.Close()

		c.HTML(302, "add_collection.html", gin.H{
			"books": books,
		})
	}

	c.Redirect(302, "/signin")
}

type BCSL struct {
	BookId int64
	Cover  string
}

type PCS struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Books       []int64 `json:"id"`
}

func PostNewCollection(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("select id from user where email = ?", session.Get("email"))
		CheckError(err)

		var userId int64

		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)

			userIdString := fmt.Sprintf("%v", userId)
			fmt.Println("User id: " + userIdString)
		}
		rows.Close()

		pc := PCS{}
		err = c.BindJSON(&pc)
		CheckError(err)

		rows, err = db.Query("select cover from book where id = ?", pc.Books[len(pc.Books)-1])
		CheckError(err)

		var cover string
		if rows.Next() {
			err := rows.Scan(&cover)
			CheckError(err)
		}
		fmt.Println(cover)
		rows.Close()

		var books string
		for i, num := range pc.Books {
			if i == (len(pc.Books) - 1) {
				books += strconv.Itoa(int(num))
				break
			}
			books += strconv.Itoa(int(num)) + ","
		}
		fmt.Println(books)

		// ----------------------------------------------
		// Fields: id, title, description, books, user_id
		// ----------------------------------------------
		stmt, err := db.Prepare("INSERT INTO `collection` (title, description, books, cover, user_id) VALUES (?, ?, ?, ?, ?)")
		CheckError(err)

		res, err := stmt.Exec(pc.Title, pc.Description, books, cover, userId)
		CheckError(err)

		id, err := res.LastInsertId()
		CheckError(err)

		fmt.Println(id)

		db.Close()

		c.String(200, strconv.Itoa(int(id)))
	}
	c.Redirect(302, "/signin")
}

func GetCollection(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("email") != nil {
		fmt.Println(session.Get("email"))

		id := c.Param("id")
		fmt.Println(id)

		db, err := sql.Open("sqlite3", "./libreread.db")
		CheckError(err)

		rows, err := db.Query("SELECT `id` FROM `user` WHERE `email` = ?", session.Get("email"))
		CheckError(err)

		var userId int64
		if rows.Next() {
			err := rows.Scan(&userId)
			CheckError(err)
		}
		fmt.Println(userId)
		rows.Close()

		rows, err = db.Query("select title, description, books from collection where id = ?", id)
		CheckError(err)

		var (
			title       string
			description string
			books       string
		)
		if rows.Next() {
			err := rows.Scan(&title, &description, &books)
			CheckError(err)
		}
		fmt.Println(books)
		rows.Close()

		bookSplit := strings.Split(books, ",")
		fmt.Println(bookSplit)

		b := []BS{}
		for i := len(bookSplit) - 1; i >= 0; i-- {
			fmt.Println(bookSplit[i])
			bookInt, err := strconv.Atoi(bookSplit[i])
			CheckError(err)

			rows, err := db.Query("select title, url, cover from book where id = ?", bookInt)
			CheckError(err)

			if rows.Next() {
				var (
					title string
					url   string
					cover string
				)
				err := rows.Scan(&title, &url, &cover)
				CheckError(err)

				b = append(b, BS{
					Title: title,
					URL:   url,
					Cover: cover,
				})
			}
			rows.Close()
		}
		fmt.Println(b)
		db.Close()

		booksList := []BSList{}
		for i := 0; i < len(b); i += 6 {
			j := i + 6
			for j > len(b) {
				j -= 1
			}
			booksList = append(booksList, b[i:j])
		}

		booksListMedium := []BSList{}
		for i := 0; i < len(b); i += 3 {
			j := i + 3
			for j > len(b) {
				j -= 1
			}
			booksListMedium = append(booksListMedium, b[i:j])
		}

		booksListSmall := []BSList{}
		for i := 0; i < len(b); i += 2 {
			j := i + 2
			for j > len(b) {
				j -= 1
			}
			booksListSmall = append(booksListSmall, b[i:j])
		}

		booksListXtraSmall := b

		c.HTML(302, "collection_item.html", gin.H{
			"title":              title,
			"description":        description,
			"booksList":          booksList,
			"booksListMedium":    booksListMedium,
			"booksListSmall":     booksListSmall,
			"booksListXtraSmall": booksListXtraSmall,
		})
	}
	c.Redirect(302, "/signin")
}
