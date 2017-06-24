/*
Copyright 2017 Nirmal Kumar

This file is part of LibreRead.

LibreRead is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

LibreRead is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with LibreRead.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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

type Env struct {
	db *sql.DB
}

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

	// Close sqlite3 database when all the functions are done
	defer db.Close()

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
	// --------------------------------------------------------------------------------------------------
	// Fields: id, title, filename, author, url, cover, pages, current_page, format, uploaded_on, user_id
	// --------------------------------------------------------------------------------------------------
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS `book` (`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		" `title` VARCHAR(255) NOT NULL, `filename` VARCHAR(255) NOT NULL," +
		" `author` VARCHAR(255) NOT NULL, `url` VARCHAR(255) NOT NULL," +
		" `cover` VARCHAR(255) NOT NULL, `pages` INTEGER NOT NULL, `current_page` INTEGER DEFAULT 0," +
		" `format` VARCHAR(255) NOT NULL, `uploaded_on` VARCHAR(255) NOT NULL, `user_id` INTEGER NOT NULL)")
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

	type Attachment struct {
		Field        string `json:"field"`
		IndexedChars int64  `json:"indexed_chars"`
	}

	type Processors struct {
		Attachment Attachment `json:"attachment"`
	}

	type AttachmentStruct struct {
		Description string       `json:"description"`
		Processors  []Processors `json:"processors"`
	}

	// Init Elasticsearch attachment
	attachment := &AttachmentStruct{
		Description: "Process documents",
		Processors: []Processors{
			Processors{
				Attachment: Attachment{
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

	type Settings struct {
		NumberOfShards   int64 `json:"number_of_shards"`
		NumberOfReplicas int64 `json:"number_of_replicas"`
	}

	type IndexStruct struct {
		Settings Settings `json:"settings"`
	}

	// Init Elasticsearch index
	index := &IndexStruct{
		Settings{
			NumberOfShards:   4,
			NumberOfReplicas: 0,
		},
	}

	b, err = json.Marshal(index)
	CheckError(err)
	fmt.Println(b)

	PutJSON("http://localhost:9200/lr_index", b)

	// Set database environment
	env := &Env{db: db}
	// Router
	r.GET("/", env.GetHomePage)
	r.GET("/signin", GetSignIn)
	r.POST("/signin", env.PostSignIn)
	r.GET("/signup", GetSignUp)
	r.POST("/signup", env.PostSignUp)
	r.GET("/confirm-email", env.ConfirmEmail)
	r.GET("/new-token", env.SendNewToken)
	r.GET("/signout", GetSignOut)
	r.POST("/upload", env.UploadBook)
	r.GET("/book/:bookname", env.SendBook)
	r.GET("/cover/:covername", SendBookCover)
	r.GET("/books/:pagination", env.GetPagination)
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

func _GetEmailFromSession(c *gin.Context) interface{} {
	session := sessions.Default(c)
	return session.Get("email")
}

func (e *Env) _GetUserId(email string) int64 {
	rows, err := e.db.Query("SELECT `id` FROM `user` WHERE `email` = ?", email)
	CheckError(err)

	var userId int64
	if rows.Next() {
		err := rows.Scan(&userId)
		CheckError(err)
	}
	rows.Close()

	return userId
}

func (e *Env) _GetBookId(fileName string) int64 {
	rows, err := e.db.Query("SELECT `id` FROM `book` WHERE `filename` = ?", fileName)
	CheckError(err)

	var bookId int64
	if rows.Next() {
		err := rows.Scan(&bookId)
		CheckError(err)
	}
	rows.Close()

	return bookId
}

func (e *Env) _CheckCurrentlyReading(bookId int64) int64 {
	rows, err := e.db.Query("SELECT `id` FROM `currently_reading` WHERE `book_id` = ?", bookId)
	CheckError(err)

	var currentlyReadingId int64
	if rows.Next() {
		err := rows.Scan(&currentlyReadingId)
		CheckError(err)
	}
	rows.Close()

	return currentlyReadingId
}

func (e *Env) _UpdateCurrentlyReading(currentlyReadingId int64, bookId int64, userId int64, dateRead string) {
	if currentlyReadingId == 0 {
		// Insert a new record
		stmt, err := e.db.Prepare("INSERT INTO `currently_reading` (book_id, user_id, date_read) VALUES (?, ?, ?)")
		CheckError(err)

		res, err := stmt.Exec(bookId, userId, dateRead)
		CheckError(err)

		id, err := res.LastInsertId()
		CheckError(err)

		fmt.Println(id)
	} else {
		// Update dateRead for the given currentlyReadingId
		stmt, err := e.db.Prepare("UPDATE `currently_reading` SET date_read=? WHERE id=?")
		CheckError(err)

		_, err = stmt.Exec(dateRead, currentlyReadingId)
		CheckError(err)
	}
}

func _GetCurrentTime() string {
	t := time.Now()
	return t.Format("20060102150405")
}

func (e *Env) SendBook(c *gin.Context) {
	email := _GetEmailFromSession(c)
	if email != nil {
		name := c.Param("bookname")

		// Get user id
		userId := e._GetUserId(email.(string))

		// Get book id
		bookId := e._GetBookId(name)

		// Get current time for date read to be used for currently reading feature
		dateRead := _GetCurrentTime()

		// Check if book already exists in currently_reading table
		currentlyReadingId := e._CheckCurrentlyReading(bookId)

		// Update currently_reading table
		e._UpdateCurrentlyReading(currentlyReadingId, bookId, userId, dateRead)

		// Return viewer.html for PDF viewer
		c.HTML(200, "viewer.html", "")
	}

	// if not signed in, redirect to sign in page
	c.Redirect(302, "/signin")
}

func SendBookCover(c *gin.Context) {
	name := c.Param("covername")
	filePath := "./uploads/img/" + name

	c.File(filePath)
}

type QuoteStruct struct {
	Author      string
	AuthorURL   string
	FromBook    string
	FromBookURL string
	Image       string
	Quote       string
}

type BookStruct struct {
	Title string
	URL   string
	Cover string
}

type BookStructList []BookStruct

func (e *Env) _GetCurrentlyReadingBooks(userId int64) []int64 {
	rows, err := e.db.Query("SELECT `book_id` FROM `currently_reading` WHERE `user_id` = ? ORDER BY `date_read` DESC LIMIT ?, ?", userId, 0, 12)
	CheckError(err)

	var crBooks []int64
	for rows.Next() {
		var crBook int64
		err = rows.Scan(&crBook)
		CheckError(err)

		crBooks = append(crBooks, crBook)
	}
	rows.Close()

	return crBooks
}

func (e *Env) _GetBook(bookId int64) (string, string, string) {
	rows, err := e.db.Query("SELECT `title`, `url`, `cover` FROM `book` WHERE `id` = ?", bookId)
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
	}
	rows.Close()

	return title, url, cover
}

func (e *Env) _GetTotalBooksCount(userId int64) int64 {
	rows, err := e.db.Query("SELECT COUNT(*) AS count FROM `book` WHERE `user_id` = ?", userId)
	CheckError(err)

	var count int64
	for rows.Next() {
		err = rows.Scan(&count)
		CheckError(err)
	}
	rows.Close()

	return count
}

func _GetTotalPages(booksCount int64) int64 {
	totalPagesFloat := float64(float64(booksCount) / 18.0)
	totalPagesDecimal := fmt.Sprintf("%.1f", totalPagesFloat)

	var totalPages int64
	if strings.Split(totalPagesDecimal, ".")[1] == "0" {
		totalPages = int64(totalPages)
	} else {
		totalPages = int64(totalPages) + 1
	}

	return totalPages
}

func (e *Env) _GetPaginatedBooks(userId int64, startIndex int64, endIndex int64) *BookStructList {
	rows, err := e.db.Query("SELECT `title`, `url`, `cover` FROM `book` WHERE `user_id` = ? ORDER BY `id` DESC LIMIT ?, ?", userId, startIndex, endIndex)
	CheckError(err)

	books := BookStructList{}

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

		books = append(books, BookStruct{
			title,
			url,
			cover,
		})
	}
	rows.Close()

	return &books
}

func _ConstructBooksWithCount(books *BookStructList, length int64) []BookStructList {
	booksList := []BookStructList{}
	var i, j int64
	for i = 0; i < int64(len(*books)); i += length {
		j = i + length
		for j > int64(len(*books)) {
			j -= 1
		}
		booksList = append(booksList, (*books)[i:j])
	}

	return booksList
}

func (e *Env) _ConstructBooksForPagination(userId int64) (int64, *BookStructList, []BookStructList, []BookStructList, []BookStructList) {
	// Check total number of rows in book table
	booksCount := e._GetTotalBooksCount(userId)

	// With Total Books count, Get Total pages required
	totalPages := _GetTotalPages(booksCount)

	// Get first 18 books for the home page
	books := e._GetPaginatedBooks(userId, 0, 18)

	// Construct books of length 6 for large screen size
	booksList := _ConstructBooksWithCount(books, 6)

	// Construct books of length 3 for medium screen size
	booksListMedium := _ConstructBooksWithCount(books, 3)

	// Construct books of length 2 for small screen size
	booksListSmall := _ConstructBooksWithCount(books, 2)

	return totalPages, books, booksList, booksListMedium, booksListSmall
}

func (e *Env) GetHomePage(c *gin.Context) {
	// Get session from cookie. Check if email exists
	// show Home page else redirect to signin page.
	email := _GetEmailFromSession(c)
	if email != nil {
		q := QuoteStruct{}
		GetJSON("http://localhost:3000/quote-of-the-day", q)

		if q.Quote == "" {
			q.Quote = "So many things are possible just as long as you don't know they're impossible."
			q.Author = "Norton Juster"
			q.AuthorURL = "https://www.goodreads.com/author/show/214.Norton_Juster"
			q.Image = "https://images.gr-assets.com/authors/1201117378p5/214.jpg"
			q.FromBook = "The Phantom Tollbooth"
			q.FromBookURL = "https://www.goodreads.com/work/1782584"
		}

		userId := e._GetUserId(email.(string))

		// Get currently reading books.
		crBooks := e._GetCurrentlyReadingBooks(userId)

		// Get book title, url, cover for currently reading books.
		currentlyReadingBooks := BookStructList{}
		for _, bookId := range crBooks {
			title, url, cover := e._GetBook(bookId)
			currentlyReadingBooks = append(currentlyReadingBooks, BookStruct{
				title,
				url,
				cover,
			})
		}

		totalPages, books, booksList, booksListMedium, booksListSmall := e._ConstructBooksForPagination(userId)

		c.HTML(302, "index.html", gin.H{
			"q": q,
			"currentlyReadingBooks": currentlyReadingBooks,
			"booksList":             booksList,
			"booksListMedium":       booksListMedium,
			"booksListSmall":        booksListSmall,
			"booksListXtraSmall":    *books,
			"totalPages":            totalPages,
		})
	}
	c.Redirect(302, "/signin")
}

func (e *Env) GetPagination(c *gin.Context) {
	email := _GetEmailFromSession(c)
	if email != nil {
		pagination, err := strconv.Atoi(c.Param("pagination"))
		CheckError(err)

		userId := e._GetUserId(email.(string))

		totalPages, books, booksList, booksListMedium, booksListSmall := e._ConstructBooksForPagination(userId)

		c.HTML(302, "pagination.html", gin.H{
			"pagination":         pagination,
			"booksList":          booksList,
			"booksListMedium":    booksListMedium,
			"booksListSmall":     booksListSmall,
			"booksListXtraSmall": books,
			"totalPages":         totalPages,
		})
	}
	c.Redirect(302, "/signin")
}

func GetSignIn(c *gin.Context) {
	email := _GetEmailFromSession(c)
	if email != nil {
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

func (e *Env) _GetHashedPassword(email string) []byte {
	rows, err := e.db.Query("select password_hash from user where email = ?", email)
	CheckError(err)

	var hashedPassword []byte

	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&hashedPassword)
		CheckError(err)
	}

	return hashedPassword
}

func _CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	// Comparing the password with the hash
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	return err
}

func (e *Env) PostSignIn(c *gin.Context) {
	email := c.PostForm("email")
	password := []byte(c.PostForm("password"))

	hashedPassword := e._GetHashedPassword(email)

	err := _CompareHashAndPassword(hashedPassword, password)

	// err nil means it is a match
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

func (e *Env) PostSignUp(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := []byte(c.PostForm("password"))

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	CheckError(err)

	stmt, err := e.db.Prepare("INSERT INTO user (name, email, password_hash) VALUES (?, ?, ?)")
	CheckError(err)

	res, err := stmt.Exec(name, email, hashedPassword)
	CheckError(err)

	id, err := res.LastInsertId()
	CheckError(err)

	go e._SendConfirmationEmail(int64(id), name, email)

	c.HTML(302, "confirm_email.html", "")

}

// For confirm email token
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int64) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (e *Env) _FillConfirmTable(token string, dateGenerated string, dateExpires string, userId int64) {
	stmt, err := e.db.Prepare("INSERT INTO confirm (token, date_generated, date_expires, user_id) VALUES (?, ?, ?, ?)")
	CheckError(err)

	_, err = stmt.Exec(token, dateGenerated, dateExpires, userId)
	CheckError(err)
}

func _SendEmail(token string, email string, name string) {
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

func (e *Env) _SendConfirmationEmail(userId int64, name string, email string) {

	// Set home many CPU cores this function wants to use
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())

	token := RandSeq(40)

	dateGenerated := _GetCurrentTime()

	// Apply one month time for token expiry
	t := time.Now()
	dateExpires := t.AddDate(0, 1, 0).Format("20060102150405")

	e._FillConfirmTable(token, dateGenerated, dateExpires, userId)

	_SendEmail(token, email, name)
}

func (e *Env) _GetConfirmTableRecord(token string, c *gin.Context) (int64, string, int64) {
	// Get id from confirm table with the token got from url.
	rows, err := e.db.Query("select id, date_expires, user_id from confirm where token = ?", token)
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
		return 0, "", 0
	}
	rows.Close()

	return id, dateExpires, userId
}

func (e *Env) _UpdateConfirmTable(currentDateTime string, used int64, id int64) {
	stmt, err := e.db.Prepare("update confirm set date_used=?, used=? where id=?")
	CheckError(err)

	_, err = stmt.Exec(currentDateTime, 1, id)
	CheckError(err)
}

func (e *Env) _SetUserConfirmed(confirmed int64, userId int64) {
	stmt, err := e.db.Prepare("update user set confirmed=? where id=?")
	CheckError(err)

	_, err = stmt.Exec(confirmed, userId)
	CheckError(err)
}

func (e *Env) ConfirmEmail(c *gin.Context) {
	token := c.Request.URL.Query()["token"][0]

	id, dateExpires, userId := e._GetConfirmTableRecord(token, c)

	if currentDateTime := _GetCurrentTime(); currentDateTime < dateExpires {
		e._UpdateConfirmTable(currentDateTime, 1, id)

		e._SetUserConfirmed(1, userId)

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

func (e *Env) _GetNameEmailFromUser(userId int64) (string, string) {
	rows, err := e.db.Query("select name, email from user where id = ?", userId)
	CheckError(err)

	var (
		name  string
		email string
	)

	if rows.Next() {
		err := rows.Scan(&name, &email)
		CheckError(err)
	}
	rows.Close()

	return name, email
}

func (e *Env) SendNewToken(c *gin.Context) {
	userId, err := strconv.Atoi(c.Request.URL.Query()["id"][0])
	CheckError(err)

	name, email := e._GetNameEmailFromUser(int64(userId))

	go e._SendConfirmationEmail(int64(userId), name, email)
}

func _ConstructFileNameForBook(fileName string, contentType string) string {
	if contentType == "application/pdf" {
		fileName = strings.Split(fileName, ".pdf")[0]
		fileName = strings.Join(strings.Split(fileName, " "), "_") + ".pdf"
	} else if contentType == "application/epub+zip" {
		fileName = strings.Split(fileName, ".epub")[0]
		fileName = strings.Join(strings.Split(fileName, " "), "_") + ".epub"
	}

	return fileName
}

func _HasPrefix(opSplit []string, content string) string {
	for _, element := range opSplit {
		if strings.HasPrefix(element, content) {
			return strings.Trim(strings.Split(element, ":")[1], " ")
		}
	}
	return ""
}

func _GetPDFInfo(filePath string) (string, string, string) {
	cmd := exec.Command("pdfinfo", filePath)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	CheckError(err)

	output := out.String()
	opSplit := strings.Split(output, "\n")

	// Get book title.
	title := _HasPrefix(opSplit, "Title")

	// Get author of the uploaded book.
	author := _HasPrefix(opSplit, "Author")

	// Get total number of pages.
	pages := _HasPrefix(opSplit, "Pages")

	return title, author, pages
}

func _GeneratePDFCover(fileName, filePath, coverPath string) string {
	cmd := exec.Command("pdfimages", "-p", "-png", "-f", "1", "-l", "2", filePath, coverPath)

	err := cmd.Run()
	CheckError(err)

	if _, err := os.Stat(coverPath + "-001-000.png"); err == nil {
		cover := "/cover/" + fileName + "-001-000.png"
		return cover
	}
	return ""
}

func (e *Env) _InsertBookRecord(
	title string,
	fileName string,
	author string,
	url string,
	cover string,
	pagesInt int64,
	format string,
	uploadedOn string,
	userId int64,
) int64 {
	stmt, err := e.db.Prepare("INSERT INTO book (title, filename, author, url, cover, pages, format, uploaded_on, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	CheckError(err)

	res, err := stmt.Exec(title, fileName, author, url, cover, pagesInt, "pdf", uploadedOn, userId)
	CheckError(err)

	id, err := res.LastInsertId()
	CheckError(err)

	return id
}

type BookInfoStruct struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	URL    string `json:"url"`
	Cover  string `json:"cover"`
}

func _PDFSeparate(path string, filePath string, wg *sync.WaitGroup) error {
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

func _ConstructPDFIndexURL(userId int64, bookId int64, i int64, pageJSON []byte) {
	indexURL := "http://localhost:9200/lr_index/book_detail/" +
		strconv.Itoa(int(userId)) + "_" + strconv.Itoa(int(bookId)) +
		"_" + strconv.Itoa(int(i)) + "?pipeline=attachment"
	fmt.Println("Index URL: " + indexURL)
	PutJSON(indexURL, pageJSON)
}

type BookDataStruct struct {
	TheData string `json:"thedata"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	URL     string `json:"url"`
	Cover   string `json:"cover"`
	Page    int64  `json:"page"`
}

func _LoopThroughSplittedPages(userId int64, bookId int64, pagesInt int64, splitPDFPath string, title string, author string, url string, cover string) {
	var i int64
	for i = 1; i < (pagesInt + 1); i += 1 {
		pagePath := splitPDFPath + "/" + strconv.Itoa(int(i)) + ".pdf"
		if _, err := os.Stat(pagePath); os.IsNotExist(err) {
			continue
		}
		data, err := ioutil.ReadFile(pagePath)
		CheckError(err)

		sEnc := base64.StdEncoding.EncodeToString([]byte(string(data)))

		bookDetail := BookDataStruct{
			TheData: sEnc,
			Title:   title,
			Author:  author,
			URL:     url,
			Cover:   cover,
			Page:    i,
		}

		pageJSON, err := json.Marshal(bookDetail)
		CheckError(err)

		_ConstructPDFIndexURL(userId, bookId, i, pageJSON)
	}
}

func FeedPDFContent(filePath string, userId int64, bookId int64, title string, author string, url string, cover string, pagesInt int64) {
	// Set home many CPU cores this function wants to use.
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())

	timeNow := _GetCurrentTime()
	splitPDFPath := "./uploads/splitpdf_" + strconv.Itoa(int(userId)) + "_" + timeNow
	if _, err := os.Stat(splitPDFPath); os.IsNotExist(err) {
		os.Mkdir(splitPDFPath, 0700)
	}

	defer os.RemoveAll(splitPDFPath)

	var wg sync.WaitGroup
	wg.Add(1)
	go _PDFSeparate(splitPDFPath, filePath, &wg)
	wg.Wait()
	fmt.Println("wg done!")

	_LoopThroughSplittedPages(userId, bookId, pagesInt, splitPDFPath, title, author, url, cover)
}

func _EPUBUnzip(filePath string, fileName string) error {
	cmd := exec.Command("unzip", filePath, "-d", "uploads/"+fileName+"/")

	err := cmd.Start()
	CheckError(err)
	fmt.Println("Waiting for command to finish...")
	err = cmd.Wait()
	fmt.Printf("Command finished with error: %v", err)
	return nil
}

// struct for META-INF/container.xml

type XMLContainerStruct struct {
	RootFiles XMLRootFiles `xml:"rootfiles"`
}

type XMLRootFiles struct {
	RootFile XMLRootFile `xml:"rootfile"`
}

type XMLRootFile struct {
	FullPath string `xml:"full-path,attr"`
}

// struct for package.xhtml derived from package.opf

type OPFMetadataStruct struct {
	Metadata OPFMetadata `xml:"metadata"`
	Spine    OPFSpine    `xml:"spine"`
	Manifest OPFManifest `xml:"manifest"`
}

type OPFMetadata struct {
	Title  string `xml:"title"`
	Author string `xml:"creator"`
}

type OPFSpine struct {
	ItemRef OPFItemRef `xml:"itemref"`
}

type OPFItemRef struct {
	IdRef []string `xml:"idref,attr"`
}

type OPFManifest struct {
	Item OPFItem `xml:"item"`
}

type OPFItem struct {
	Id        []string `xml:"id,attr"`
	Href      []string `xml:"href,attr"`
	MediaType []string `xml:"media-type,attr"`
}

func (e *Env) UploadBook(c *gin.Context) {
	email := _GetEmailFromSession(c)
	if email != nil {
		userId := e._GetUserId(email.(string))

		multipart, err := c.Request.MultipartReader()
		CheckError(err)

		for {
			mimePart, err := multipart.NextPart()

			if err == io.EOF {
				break
			}

			// Get filename and content type
			_, params, err := mime.ParseMediaType(mimePart.Header.Get("Content-Disposition"))
			CheckError(err)
			contentType, _, err := mime.ParseMediaType(mimePart.Header.Get("Content-Type"))
			CheckError(err)

			// Construct filename for the book uploaded
			fileName := _ConstructFileNameForBook(params["filename"], contentType)

			bookId := e._GetBookId(fileName)

			if bookId != 0 {
				c.String(200, fileName+" already exists. ")
				continue
			}

			uploadedOn := _GetCurrentTime()

			filePath := "./uploads/" + fileName

			out, err := os.Create(filePath)
			CheckError(err)

			_, err = io.Copy(out, mimePart)
			CheckError(err)

			out.Close()

			if contentType == "application/pdf" {

				title, author, pages := _GetPDFInfo(filePath)

				if title == "" {
					title = fileName
				}

				if author == "" {
					author = "unknown"
				}

				fmt.Println("Book title: " + title)
				fmt.Println("Book author: " + author)
				fmt.Println("Total pages: " + pages)

				pagesInt, err := strconv.ParseInt(pages, 10, 64)
				CheckError(err)

				url := "/book/" + fileName
				fmt.Println("Book URL: " + url)

				coverPath := "./uploads/img/" + fileName

				cover := _GeneratePDFCover(fileName, filePath, coverPath)

				fmt.Println("Book cover: " + cover)

				// Insert new book in `book` table
				bookId := e._InsertBookRecord(title, fileName, author, url, cover, pagesInt, "pdf", uploadedOn, userId)

				// Feed book info to ES
				bookInfo := BookInfoStruct{
					Title:  title,
					Author: author,
					URL:    url,
					Cover:  cover,
				}

				fmt.Println(bookInfo)

				indexURL := "http://localhost:9200/lr_index/book_info/" + strconv.Itoa(int(bookId))
				fmt.Println(indexURL)

				b, err := json.Marshal(bookInfo)
				CheckError(err)

				PutJSON(indexURL, b)

				// Feed book content to ES
				go FeedPDFContent(filePath, userId, bookId, title, author, url, cover, pagesInt)

				c.String(200, fileName+" uploaded successfully. ")

			} else if contentType == "application/epub+zip" {

			}
		}
	}
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
	hitsBIS := []BookInfoStruct{}
	for _, el := range hits {
		hitsBIS = append(hitsBIS, BookInfoStruct{
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
	BookInfo   []BookInfoStruct `json:"book_info"`
	BookDetail []BDRSHS         `json:"book_detail"`
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
	Source BookInfoStruct `json:"_source"`
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

		b := BookStructList{}
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

				b = append(b, BookStruct{
					Title: title,
					URL:   url,
					Cover: cover,
				})
			}
			rows.Close()
		}
		fmt.Println(b)
		db.Close()

		booksList := []BookStructList{}
		for i := 0; i < len(b); i += 6 {
			j := i + 6
			for j > len(b) {
				j -= 1
			}
			booksList = append(booksList, b[i:j])
		}

		booksListMedium := []BookStructList{}
		for i := 0; i < len(b); i += 3 {
			j := i + 3
			for j > len(b) {
				j -= 1
			}
			booksListMedium = append(booksListMedium, b[i:j])
		}

		booksListSmall := []BookStructList{}
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
